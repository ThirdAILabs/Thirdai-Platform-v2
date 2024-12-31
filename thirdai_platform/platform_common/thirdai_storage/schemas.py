from sqlalchemy import Boolean, Column, DateTime, ForeignKey, Integer, String, func
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import relationship

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


class XMLLog(Base):
    __tablename__ = "xml_log"
    id = Column(String, primary_key=True, index=True)
    xml_string = Column(String)
    timestamp = Column(DateTime, default=func.current_timestamp())

    # Using association tables for many-to-many relationships
    elements = relationship(
        "XMLElement", secondary="log_element_association", back_populates="logs"
    )
    feedback = relationship(
        "XMLFeedback", secondary="log_feedback_association", back_populates="logs"
    )


class LogElementAssociation(Base):
    __tablename__ = "log_element_association"
    log_id = Column(String, ForeignKey("xml_log.id"), primary_key=True)
    element_id = Column(Integer, ForeignKey("xml_element.id"), primary_key=True)


class LogFeedbackAssociation(Base):
    __tablename__ = "log_feedback_association"
    log_id = Column(String, ForeignKey("xml_log.id"), primary_key=True)
    feedback_id = Column(Integer, ForeignKey("xml_feedback.id"), primary_key=True)


class XMLElement(Base):
    __tablename__ = "xml_element"
    id = Column(Integer, primary_key=True, autoincrement=True)
    xpath = Column(String, index=True)
    attribute = Column(String, index=True, nullable=True)
    n_tokens = Column(Integer, index=True)

    # Many-to-many relationship with logs
    logs = relationship(
        "XMLLog", secondary="log_element_association", back_populates="elements"
    )


class XMLFeedback(Base):
    __tablename__ = "xml_feedback"
    id = Column(Integer, primary_key=True, autoincrement=True)
    xpath = Column(String, index=True)
    attribute = Column(String, index=True, nullable=True)
    token_start = Column(Integer)
    token_end = Column(Integer)
    n_tokens = Column(Integer, index=True)
    label = Column(String)
    user_provided = Column(Boolean, default=False)
    status = Column(String)

    # Many-to-many relationship with logs
    logs = relationship(
        "XMLLog", secondary="log_feedback_association", back_populates="feedback"
    )
