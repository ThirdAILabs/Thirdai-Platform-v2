from __future__ import annotations

import typing
from abc import abstractmethod
from collections import defaultdict

from sqlalchemy import create_engine, func
from sqlalchemy.orm import scoped_session, sessionmaker

from data_types import DataType, deserialize_datatype
from schemas import Base, LogEntry


class DataStore:
    # Interface for data store backend. Can be repurposed to a DB based storage,
    # a file based storage, etc. The DataStore should be persistent.
    @abstractmethod
    def add_entries(self, entries: typing.List[typing.tuple[str, str, str]]):
        # batch insertion of entries to the store
        pass

    @abstractmethod
    def get_count(self, name: str):
        # returns the total count of entries with the given name
        pass

    @abstractmethod
    def delete_oldest_entries(self, name: str, num_entries: int):
        # sort the entries by timestamp and deletes the oldest k entries
        pass

    @abstractmethod
    def get_entries(self, name: str, num_entries: int):
        # get num_entries entries with a specific name
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

    def add_entries(self, entries: typing.List[typing.Tuple[str, str, str]]):
        session = self.Session()
        session.bulk_insert_mappings(
            LogEntry,
            [
                {"datatype": datatype, "name": name, "serialized_data": data}
                for datatype, name, data in entries
            ],
        )
        session.commit()

    def get_count(self, name: str):
        session = self.Session()
        count = (
            session.query(func.count(LogEntry.id))
            .filter(LogEntry.name == name)
            .scalar()
        )
        return count

    def delete_oldest_entries(self, name: str, num_entries: int):
        session = self.Session()
        oldest_entries = (
            session.query(LogEntry.id)
            .filter(LogEntry.name == name)
            .order_by(LogEntry.timestamp.asc())
            .limit(num_entries)
            .all()
        )
        for entry_id in oldest_entries:
            session.delete(
                session.query(LogEntry).filter(LogEntry.id == entry_id[0]).one()
            )
        session.commit()

    def get_entries(self, name: str, num_entries: int):
        session = self.Session()
        entries = (
            session.query(LogEntry.datatype, LogEntry.serialized_data)
            .filter(LogEntry.name == name)
            .order_by(LogEntry.timestamp.desc())
            .limit(num_entries)
            .all()
        )
        return entries

    def existing_names(self):
        session = self.Session()
        names = session.query(LogEntry.name).distinct().all()

        return set([name[0] for name in names])


class DataStorage:
    def __init__(self, connector: DataStore, per_log_buffer_size: int = 1000):
        self.connector = connector

        # class attributes are generated using the connector hence, we do not need to write
        # save load logic for DataStorage class.
        self.existing_entities: typing.Set[str] = connector.existing_names()
        self.sample_counter = defaultdict(int)

        for name in self.existing_entities:
            self.sample_counter[name] = connector.get_count(name=name)

        # if per log buffer size is None then no limit on the number of samples for any logtype
        self.per_log_buffer_size = per_log_buffer_size

    def insert(self, samples: typing.List[DataType]):

        samples_to_insert = []
        for sample in samples:

            self.existing_entities.add(sample.name)

            if (
                self.per_log_buffer_size
                and self.sample_counter[sample.name] < self.per_log_buffer_size
            ):
                samples_to_insert.append(
                    (sample.datatype, sample.name, sample.serialize())
                )

                self.sample_counter[sample.name] += 1

        self.connector.add_entries(samples_to_insert)

    def retrieve(self, name: str, num_samples: int):
        if name not in self.existing_entities:
            return []

        entries = self.connector.get_entries(name, num_entries=num_samples)

        return [
            deserialize_datatype(type=datatype, name=name, serialized_data=data)
            for datatype, data in entries
        ]
