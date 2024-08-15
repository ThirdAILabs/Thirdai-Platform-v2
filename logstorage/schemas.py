from sqlalchemy import Column, DateTime, Integer, String, create_engine, func
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()


class LogEntry(Base):
    __tablename__ = "entries"
    id = Column(Integer, primary_key=True)
    datatype = Column(String)
    name = Column(String, index=True)
    serialized_data = Column(String)
    timestamp = Column(DateTime, default=func.current_timestamp())
