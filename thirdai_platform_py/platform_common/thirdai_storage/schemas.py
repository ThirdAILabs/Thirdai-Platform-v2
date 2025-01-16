from sqlalchemy import Boolean, Column, DateTime, Integer, String, func
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()


class Samples(Base):
    __tablename__ = "samples"
    id = Column(String(36), primary_key=True, index=True)
    datatype = Column(String)
    name = Column(String, index=True)
    status = Column(String)
    serialized_data = Column(String)
    user_provided = Column(Boolean, nullable=False, default=False)
    timestamp = Column(DateTime, default=func.current_timestamp())


class SampleSeen(Base):
    # this is used for reservoir sampling for generating balancing samples
    __tablename__ = "sample_seen"
    name = Column(String, primary_key=True, index=True)
    seen = Column(Integer, default=0)


class MetaData(Base):
    __tablename__ = "metadata"
    name = Column(String, primary_key=True, index=True)
    datatype = Column(String)
    serialized_data = Column(String)
    status = Column(String)
