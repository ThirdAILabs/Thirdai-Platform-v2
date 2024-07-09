FROM tiangolo/uvicorn-gunicorn-fastapi:python3.9

WORKDIR /app

# Install dependencies before copying the directory so python dependencies are 
# cached even when there are changes in the directory.
RUN python3 -m pip install thirdai[neural_db]

COPY . /app

CMD ["python3", "run.py"]