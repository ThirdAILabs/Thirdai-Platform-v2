from __future__ import annotations

import typing
from abc import abstractmethod
from collections import defaultdict

from sqlalchemy import create_engine, func
from sqlalchemy.orm import scoped_session, sessionmaker

from data_types import (
    DataSamples,
    UserFeedBack,
    deserialize_sample_datatype,
    deserialize_userfeedback,
)
from schemas import Base, Samples, FeedBack


from sqlalchemy.engine import Engine
from sqlalchemy import event


@event.listens_for(Engine, "connect")
def set_sqlite_pragma(dbapi_connection, connection_record):
    cursor = dbapi_connection.cursor()
    cursor.execute("PRAGMA foreign_keys=ON")
    cursor.close()


class DataStore:
    # Interface for data store backend. Can be repurposed to a DB based storage,
    # a file based storage, etc. The DataStore should be persistent.
    @abstractmethod
    def add_samples(self, entries: typing.List[typing.tuple[str, str, str, str]]):
        # batch insertion of samples to the store
        pass

    @abstractmethod
    def add_feedback(self, entries: typing.List[typing.tuple[str, str, str]]):
        # batch insertion of user feedback to the store
        pass

    @abstractmethod
    def get_sample_count(self, name: str):
        # returns the total count of entries with the given name
        pass

    @abstractmethod
    def get_feedback_count(self, name: str):
        pass

    @abstractmethod
    def delete_old_samples(self, name: str, samples_to_store: int):
        # sort the entries by timestamp and deletes the oldest k entries
        pass

    @abstractmethod
    def get_samples(self, name: str, num_samples: int):
        # get num_entries entries with a specific name
        pass

    @abstractmethod
    def get_feedback(self, name: str):
        pass

    @abstractmethod
    def existing_names(self):
        # get unique names in the store
        pass


class SQLiteStore(DataStore):
    def __init__(self, db_path: str):
        self.engine = create_engine(f"sqlite:///{db_path}", echo=True)
        self.Session = scoped_session(sessionmaker(bind=self.engine))
        Base.metadata.create_all(self.engine)

    def add_samples(self, entries: typing.List[typing.Tuple[str, str, str, str]]):
        session = self.Session()
        session.bulk_insert_mappings(
            Samples,
            [
                {
                    "id": unique_id,
                    "datatype": datatype,
                    "name": name,
                    "serialized_data": data,
                }
                for unique_id, datatype, name, data in entries
            ],
        )
        session.commit()

    def add_feedback(self, entries: typing.List[typing.Tuple[str, str]]):
        session = self.Session()
        session.bulk_insert_mappings(
            FeedBack,
            [
                {
                    "sample_uuid": sample_uuid,
                    "serialized_data": data,
                }
                for sample_uuid, data in entries
            ],
        )
        session.commit()

    def get_sample_count(self, name: str):
        session = self.Session()
        count = (
            session.query(func.count(Samples.id)).filter(Samples.name == name).scalar()
        )
        return count

    def get_feedback_count(self, name: str):
        session = self.Session()
        count = (
            session.query(func.count(FeedBack.id))
            .join(Samples)
            .filter(Samples.name == name)
            .scalar()
        )
        return count

    def delete_old_samples(self, name: str, samples_to_store: int):
        current_count = self.get_sample_count(name)

        samples_to_delete = current_count - samples_to_store

        if samples_to_delete > 0:
            session = self.Session()

            oldest_entries = (
                session.query(Samples.id)
                .filter(Samples.name == name)
                .order_by(Samples.timestamp.asc())
                .limit(samples_to_delete)
                .all()
            )
            for entry_id in oldest_entries:
                session.delete(
                    session.query(Samples).filter(Samples.id == entry_id[0]).one()
                )
            session.commit()

    def get_samples(self, name: str, num_samples: int):
        session = self.Session()
        entries = (
            session.query(Samples.datatype, Samples.id, Samples.serialized_data)
            .filter(Samples.name == name)
            .order_by(Samples.timestamp.desc())
            .limit(num_samples)
            .all()
        )
        return entries

    def get_feedback(self, name: str):
        session = self.Session()
        entries = (
            session.query(
                Samples.datatype, FeedBack.sample_uuid, FeedBack.serialized_data
            )
            .join(Samples)
            .filter(Samples.name == name)
            .order_by(FeedBack.timestamp.desc())
            .all()
        )
        return entries

    def existing_names(self):
        session = self.Session()
        names = session.query(Samples.name).distinct().all()

        return set([name[0] for name in names])


class DataStorage:
    def __init__(self, connector: DataStore, per_log_buffer_size: int = 1000):
        self.connector = connector

        # class attributes are generated using the connector hence, we do not need to write
        # save load logic for DataStorage class.
        self.existing_entities: typing.Set[str] = connector.existing_names()
        self.sample_counter = defaultdict(int)

        for name in self.existing_entities:
            self.sample_counter[name] = connector.get_sample_count(name=name)

        # if per log buffer size is None then no limit on the number of samples for any logtype
        self.per_log_buffer_size = per_log_buffer_size

    def insert_samples(self, samples: typing.List[DataSamples]):

        samples_to_insert = []
        for sample in samples:
            self.existing_entities.add(sample.name)

            if (
                self.per_log_buffer_size
                and self.sample_counter[sample.name] < self.per_log_buffer_size
            ):
                samples_to_insert.append(
                    (sample.uuid, sample.datatype, sample.name, sample.serialize())
                )

                self.sample_counter[sample.name] += 1

        self.connector.add_samples(samples_to_insert)

    def retrieve_samples(self, name: str, num_samples: int):
        if name not in self.existing_entities:
            return []

        entries = self.connector.get_samples(name, num_samples=num_samples)

        return [
            deserialize_sample_datatype(
                type=datatype, unique_id=unique_id, name=name, serialized_data=data
            )
            for datatype, unique_id, data in entries
        ]

    def insert_feedbacks(self, feedbacks: typing.List[UserFeedBack]):
        feedbacks_to_insert = []

        for feedback in feedbacks:
            if feedback.name not in self.existing_entities:
                raise Exception("Cannot add feedback for a name that does not exist")

            feedbacks_to_insert.append((feedback.sample_uuid, feedback.serialize()))

        self.connector.add_feedback(feedbacks_to_insert)

    def retrieve_feedbacks(self, name: str):
        if name not in self.existing_entities:
            return []

        feedbacks = self.connector.get_feedback(name)

        return [
            deserialize_userfeedback(
                type=datatype, sample_uuid=sample_uuid, name=name, serialized_data=data
            )
            for datatype, sample_uuid, data in feedbacks
        ]
