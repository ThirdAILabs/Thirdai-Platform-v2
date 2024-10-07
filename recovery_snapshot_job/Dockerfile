FROM python:3.9

WORKDIR /app

# Install PostgreSQL client utilities (for pg_dump)
RUN apt-get update && apt-get install -y postgresql-client

# Install dependencies before copying the directory so python dependencies are 
# cached even when there are changes in the directory.
COPY ./requirements.txt /app/requirements.txt
RUN pip install -r /app/requirements.txt

COPY ./. /app

CMD ["python3", "run.py"]