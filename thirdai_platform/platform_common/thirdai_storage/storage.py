from __future__ import annotations

import typing
from abc import abstractmethod
from uuid import uuid4

from sqlalchemy import create_engine, func
from sqlalchemy.orm import scoped_session, sessionmaker

from .data_types import (
    DataSample,
    Metadata,
    MetadataStatus,
    SampleStatus,
    XMLElementData,
    XMLFeedbackData,
    XMLLogData,
)
from .schemas import (
    Base,
    LogElementAssociation,
    MetaData,
    Samples,
    SampleSeen,
    XMLElement,
    XMLFeedback,
    XMLLog,
)
from .utils import reservoir_sampling

RESERVOIR_RECENCY_MULTIPLIER = 1


class Connector:
    # Interface for data store backend. Can be repurposed to a DB based storage,
    # a file based storage, etc. The DataStore should be persistent.
    @abstractmethod
    def add_samples(
        self,
        entries: typing.List[typing.tuple[str, str, str, str, bool]],
        reservoir_size: int = None,
    ):
        """
        Performs batch insertion of samples into the data store using reservoir sampling.

        Reservoir sampling ensures a random sample of fixed size is maintained even as new samples
        are added over time. This is useful for keeping a representative subset of data when the
        total dataset is too large to store completely.

        Args:
            entries: List of tuples containing sample information with fields:
                unique_id (str): Unique identifier for the sample
                datatype (str): Type of the sample (e.g. "token_classification", "text_classification")
                name (str): Name/category of the sample (e.g. "ner", "sentiment")
                serialized_data (str): The sample data in serialized format
                status (str): Processing status of the sample (e.g. "untrained", "trained")
                user_provided (bool): Whether this sample was provided by a user (True) or generated (False)

            reservoir_size (int, optional): Maximum number of samples to maintain in the reservoir.
                If None, all samples will be stored without sampling.
        """
        pass

    @abstractmethod
    def get_sample_count(self, name: str):
        # returns the total count of entries with the given name
        pass

    @abstractmethod
    def delete_old_samples(self, name: str, samples_to_store: int):
        # sort the entries by timestamp and deletes the oldest k entries
        pass

    @abstractmethod
    def get_samples(self, name: str, num_samples: int, user_provided: bool):
        # get num_entries entries with a specific name
        pass

    @abstractmethod
    def existing_sample_names(self):
        # get unique names in the store
        pass

    @abstractmethod
    def insert_metadata(self, name: str, datatype: str, serialized_data: str):
        # if an entry with the value name exists, updates its serialized data,
        # else makes a new entry for the metadata in the DB.
        pass

    @abstractmethod
    def get_metadata(self, name: str):
        pass

    @abstractmethod
    def remove_untrained_samples(self, name: str):
        pass

    @abstractmethod
    def update_metadata_status(self, name: str, status: MetadataStatus):
        pass

    @abstractmethod
    def update_sample_status(self, name: str, status: SampleStatus):
        pass

    @abstractmethod
    def add_xml_log(self, xml_log: XMLLogData) -> str:
        pass

    @abstractmethod
    def store_user_xml_feedback(
        self, log_id: str, feedbacks: typing.List[XMLFeedbackData]
    ):
        pass

    @abstractmethod
    def get_xml_log_by_id(self, log_id: str) -> XMLLogData:
        pass

    @abstractmethod
    def get_user_provided_xml_feedback(self, status: SampleStatus):
        pass

    @abstractmethod
    def find_conflicting_xml_feedback(self, feedback: XMLFeedbackData):
        pass


class SQLiteConnector(Connector):
    def __init__(self, db_path: str):
        self.engine = create_engine(f"sqlite:///{db_path}", echo=False)
        self.Session = scoped_session(sessionmaker(bind=self.engine))
        Base.metadata.create_all(self.engine)

    def add_samples(
        self,
        entries: typing.List[typing.Tuple[str, str, str, str, str, bool]],
        reservoir_size: int = None,
    ):
        # NOTE : Reservoir Size restrictions will not be held if used in a multi-threaded environment

        if len(entries) == 0:
            return

        session = self.Session()

        reservoir_counter = (
            session.query(SampleSeen).filter(SampleSeen.name == entries[0][2]).first()
        )
        if not reservoir_counter:
            reservoir_counter = SampleSeen(name=entries[0][2], seen=0)
            session.add(reservoir_counter)

        if reservoir_size:
            # NOTE: This isn't true reservoir sampling as the elements are deleted in bulk before adding new ones. This implies that there is a bias towards keeping recent samples in the reservoir. The parameter RESERVOIR_RECENCY_MULTIPLIER can be tuned to control this bias.

            # find the indices of the entries to be added to the reservoir
            valid_indices = list(range(len(entries)))
            selected_indices = reservoir_sampling(
                valid_indices,
                reservoir_size,
                current_size=self.get_sample_count(entries[0][2]),
                total_items_seen=reservoir_counter.seen,
                recency_multipler=RESERVOIR_RECENCY_MULTIPLIER,
            )

            entries = [entries[i] for i in selected_indices]

            num_samples_to_remove = (
                len(selected_indices)
                + self.get_sample_count(entries[0][2])
                - reservoir_size
            )

            if num_samples_to_remove > 0:
                samples_to_remove = (
                    session.query(Samples)
                    .filter(Samples.name == entries[0][2])
                    .order_by(func.random())
                    .limit(num_samples_to_remove)
                    .all()
                )
                session.query(Samples).filter(
                    Samples.id.in_([sample.id for sample in samples_to_remove])
                ).delete(synchronize_session=False)

        session.bulk_insert_mappings(
            Samples,
            [
                {
                    "id": unique_id,
                    "datatype": datatype,
                    "name": name,
                    "serialized_data": data,
                    "status": status,
                    "user_provided": user_provided,
                }
                for unique_id, datatype, name, data, status, user_provided in entries
            ],
        )

        reservoir_counter.seen = reservoir_counter.seen + len(entries)
        session.commit()

    def get_sample_count(self, name: str):

        session = self.Session()

        return (
            session.query(func.count(Samples.id)).filter(Samples.name == name).scalar()
        )

    def delete_old_samples(self, name: str, samples_to_store: int):
        # total samples
        total_samples = self.get_sample_count(name)

        samples_to_delete = total_samples - samples_to_store

        if samples_to_delete > 0:
            session = self.Session()
            oldest_entries = (
                session.query(Samples.id)
                .filter(Samples.name == name)
                .filter(Samples.user_provided == False)
                .order_by(Samples.timestamp.asc())
                .limit(samples_to_delete)
                .all()
            )

            for entry_id in oldest_entries:
                session.delete(session.query(Samples).get(entry_id[0]))
            session.commit()

    def get_samples(self, name: str, num_samples: int, user_provided: bool):
        session = self.Session()
        entries = (
            session.query(
                Samples.datatype,
                Samples.id,
                Samples.serialized_data,
                Samples.status,
            )
            .filter(Samples.name == name)
            .filter(Samples.user_provided == user_provided)
            .order_by(Samples.timestamp.desc())
            .limit(num_samples)
            .all()
        )
        return entries

    def existing_sample_names(self):
        session = self.Session()
        names = session.query(Samples.name).distinct().all()

        return set([name[0] for name in names])

    def insert_metadata(
        self, name: str, status: str, datatype: str, serialized_data: str
    ):
        session = self.Session()

        existing_metadata = (
            session.query(MetaData).filter(MetaData.name == name).first()
        )
        if existing_metadata:
            # update the entry in place
            existing_metadata.serialized_data = serialized_data
            existing_metadata.status = status
        else:
            new_metadata = MetaData(
                name=name,
                datatype=datatype,
                serialized_data=serialized_data,
                status=status,
            )
            session.add(new_metadata)

        session.commit()

    def get_metadata(self, name: str):
        session = self.Session()

        entry = (
            session.query(
                MetaData.datatype,
                MetaData.name,
                MetaData.serialized_data,
                MetaData.status,
            )
            .filter(MetaData.name == name)
            .first()
        )
        return entry

    def remove_untrained_samples(self, name: str):
        # remove all untrained samples for a given name
        session = self.Session()
        session.query(Samples).filter(Samples.name == name).filter(
            Samples.status == SampleStatus.untrained
        ).delete()
        session.commit()

    def update_metadata_status(self, name: str, status: MetadataStatus):
        session = self.Session()
        session.query(MetaData).filter(MetaData.name == name).update(
            {MetaData.status: status}
        )
        session.commit()

    def update_sample_status(self, name: str, status: SampleStatus):
        session = self.Session()
        session.query(Samples).filter(Samples.name == name).update(
            {Samples.status: status}
        )
        session.commit()

    def add_xml_log(self, log: XMLLogData):
        session = self.Session()
        # Create XMLLog entry
        log_id = str(uuid4())
        xml_log = XMLLog(id=log_id, xml_string=log.xml_string)
        session.add(xml_log)

        # Add each element, reusing existing ones if found
        for elem in log.elements:
            element = (
                session.query(XMLElement)
                .filter_by(
                    xpath=elem.xpath,
                    attribute=elem.attribute,
                    n_tokens=elem.n_tokens,
                )
                .first()
            )

            if not element:
                element = XMLElement(
                    xpath=elem.xpath,
                    attribute=elem.attribute,
                    n_tokens=elem.n_tokens,
                )
                session.add(element)

            xml_log.elements.append(element)

        session.commit()
        return log_id

    def store_user_xml_feedback(
        self, log_id: str, feedbacks: typing.List[XMLFeedbackData]
    ):
        session = self.Session()
        xml_log = session.query(XMLLog).get(log_id)
        if not xml_log:
            raise ValueError("XML Log not found")

        for fb in feedbacks:
            existing_feedback = (
                session.query(XMLFeedback)
                .filter_by(
                    xpath=fb.xpath,
                    attribute=fb.attribute,
                    token_start=fb.token_start,
                    token_end=fb.token_end,
                    n_tokens=fb.n_tokens,
                    label=fb.label,
                )
                .first()
            )

            if existing_feedback:
                # Reuse existing feedback
                xml_log.feedback.append(existing_feedback)
            else:
                # Create new feedback only if it doesn't exist
                feedback = XMLFeedback(
                    xpath=fb.xpath,
                    attribute=fb.attribute,
                    token_start=fb.token_start,
                    token_end=fb.token_end,
                    n_tokens=fb.n_tokens,
                    label=fb.label,
                    user_provided=fb.user_provided,
                    status=fb.status,
                )
                session.add(feedback)
                xml_log.feedback.append(feedback)

        session.commit()

    def get_xml_log_by_id(self, log_id: str) -> XMLLogData:
        session = self.Session()
        log = session.query(XMLLog).get(log_id)
        return XMLLogData(
            xml_string=log.xml_string,
            elements=[
                XMLElementData(
                    xpath=elem.xpath, attribute=elem.attribute, n_tokens=elem.n_tokens
                )
                for elem in log.elements
            ],
        )

    def get_user_provided_xml_feedback(self, status: SampleStatus):
        session = self.Session()
        # Get all user provided feedback
        feedbacks = (
            session.query(XMLFeedback)
            .filter_by(user_provided=True, status=status)
            .all()
        )

        results = []
        for fb in feedbacks:
            # Find all XML logs that have matching elements
            matching_xmls = (
                session.query(XMLLog)
                .join(LogElementAssociation)
                .join(XMLElement)
                .filter(
                    XMLElement.xpath == fb.xpath,
                    XMLElement.n_tokens == fb.n_tokens,
                    XMLElement.attribute == fb.attribute,
                )
                .all()
            )

            feedback = XMLFeedbackData(
                xpath=fb.xpath,
                attribute=fb.attribute,
                token_start=fb.token_start,
                token_end=fb.token_end,
                n_tokens=fb.n_tokens,
                label=fb.label,
                user_provided=fb.user_provided,
                status=fb.status,
            )

            results.append((feedback, [log.id for log in matching_xmls]))
        return results

    def find_conflicting_xml_feedback(
        self, feedback: XMLFeedbackData
    ) -> typing.List[XMLFeedbackData]:
        session = self.Session()
        # Find any conflicting feedback
        conflicts = (
            session.query(XMLFeedback)
            .filter(
                XMLFeedback.xpath == feedback.xpath,
                XMLFeedback.attribute == feedback.attribute,
                XMLFeedback.n_tokens == feedback.n_tokens,
                XMLFeedback.token_start <= feedback.token_start,
                XMLFeedback.token_end >= feedback.token_end,
                XMLFeedback.label != feedback.label,
                XMLFeedback.status == feedback.status,
            )
            .all()
        )

        return [
            XMLFeedbackData(
                xpath=c.xpath,
                attribute=c.attribute,
                token_start=c.token_start,
                token_end=c.token_end,
                n_tokens=c.n_tokens,
                label=c.label,
                user_provided=c.user_provided,
                status=c.status,
            )
            for c in conflicts
        ]

    def update_xml_feedback_status(
        self, feedback: XMLFeedbackData, status: SampleStatus
    ):
        session = self.Session()
        session.query(XMLFeedback).filter_by(
            xpath=feedback.xpath,
            attribute=feedback.attribute,
            token_start=feedback.token_start,
            token_end=feedback.token_end,
            n_tokens=feedback.n_tokens,
            label=feedback.label,
        ).update({XMLFeedback.status: status})
        session.commit()


class DataStorage:
    def __init__(self, connector: Connector):
        # all class attributes should be generated using the connector
        # and it is supposed to be used as a single source of truth.
        self.connector = connector

        # if per name buffer size is None then no limit on the number of samples for each name
        # this attribute is set as private so that two different instances of
        # DataStorage with the same connector have same reservoir size.
        self._reservoir_size = 100000

    def insert_samples(
        self, samples: typing.List[DataSample], override_reservoir_limit=False
    ):
        samples_to_insert = []
        for sample in samples:
            samples_to_insert.append(
                (
                    sample.unique_id,
                    sample.datatype,
                    sample.name,
                    sample.serialize_data(),
                    sample.status.value,
                    sample.user_provided,
                )
            )

        self.connector.add_samples(
            samples_to_insert,
            reservoir_size=(
                self._reservoir_size if not override_reservoir_limit else None
            ),
        )

    def retrieve_samples(self, name: str, num_samples: int, user_provided: bool):
        entries = self.connector.get_samples(
            name, num_samples=num_samples, user_provided=user_provided
        )

        return [
            DataSample.from_serialized(
                type=datatype,
                unique_id=unique_id,
                name=name,
                serialized_data=data,
                status=status,
                user_provided=user_provided,
            )
            for datatype, unique_id, data, status in entries
        ]

    def clip_storage(self):
        existing_sample_types = self.connector.existing_sample_names()

        for name in existing_sample_types:
            self.connector.delete_old_samples(
                name=name, samples_to_store=self._reservoir_size
            )

    def insert_metadata(self, metadata: Metadata):
        # updates the serialized data in place if another entry with the same
        # name exists
        self.connector.insert_metadata(
            name=metadata.name,
            status=metadata.status,
            datatype=metadata.datatype,
            serialized_data=metadata.serialize_data(),
        )

    def get_metadata(self, name) -> Metadata:
        data = self.connector.get_metadata(name)
        if data:
            return Metadata.from_serialized(
                type=data[0], name=data[1], serialized_data=data[2], status=data[3]
            )

        return None

    def remove_untrained_samples(self, name: str):
        self.connector.remove_untrained_samples(name)

    def rollback_metadata(self, name: str):
        metadata = self.get_metadata(name)

        if metadata.status == MetadataStatus.updated:
            metadata.rollback()
            self.insert_metadata(metadata)

    def update_metadata_status(self, name: str, status: MetadataStatus):
        self.connector.update_metadata_status(name, status)

    def update_sample_status(self, name: str, status: SampleStatus):
        self.connector.update_sample_status(name, status)

    def add_xml_log(self, xml_log: XMLLogData) -> str:
        return self.connector.add_xml_log(xml_log)

    def store_user_xml_feedback(
        self, log_id: str, feedbacks: typing.List[XMLFeedbackData]
    ):
        self.connector.store_user_xml_feedback(log_id, feedbacks)

    def get_xml_log_by_id(self, log_id: str) -> XMLLogData:
        return self.connector.get_xml_log_by_id(log_id)

    def get_user_provided_xml_feedback(self, status: SampleStatus):
        return self.connector.get_user_provided_xml_feedback(status)

    def find_conflicting_xml_feedback(
        self, feedback: XMLFeedbackData
    ) -> typing.List[XMLFeedbackData]:
        return self.connector.find_conflicting_xml_feedback(feedback)
