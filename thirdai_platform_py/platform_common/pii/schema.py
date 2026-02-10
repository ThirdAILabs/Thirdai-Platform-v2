import json

from sqlalchemy import Column, DateTime, Integer, String, Text, create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker

Base = declarative_base()


class UDTReport(Base):
    __tablename__ = "udt_reports"

    id = Column(String, primary_key=True)
    status = Column(String, nullable=False)
    submitted_at = Column(DateTime, nullable=False)
    updated_at = Column(DateTime, nullable=False)
    attempt = Column(Integer, default=0, nullable=False)
    msg = Column(String, nullable=True, default=None)
    custom_tags = Column(Text, nullable=True, default="[]")

    def get_tags(self):
        """Return the custom tags as a list."""
        try:
            return json.loads(self.custom_tags) if self.custom_tags else []
        except Exception:
            return []

    def set_tags(self, tags_list):
        """Set the custom tags from a list."""
        self.custom_tags = json.dumps(tags_list)


def get_udt_db_session(db_path: str):
    """Dynamically create a session for the UDT database."""
    engine = create_engine(f"sqlite:///{db_path}")
    Base.metadata.create_all(bind=engine)
    SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
    return SessionLocal
