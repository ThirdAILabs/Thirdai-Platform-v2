FROM tiangolo/uvicorn-gunicorn-fastapi:python3.9

WORKDIR /app

# Install dependencies before copying the directory so python dependencies are 
# cached even when there are changes in the directory.
COPY ./requirements.txt /app/requirements.txt
RUN pip install -r /app/requirements.txt
RUN python3 -m pip install pdftitle

COPY ./thirdai-0.9.14+9d0ffa8-cp39-cp39-manylinux_2_17_x86_64.manylinux2014_x86_64.whl /app/
RUN python3 -m pip install thirdai-0.9.14+9d0ffa8-cp39-cp39-manylinux_2_17_x86_64.manylinux2014_x86_64.whl --force-reinstall

# Install thirdai package separately since this is the dependency that changes
# the most often. This allows the other packages to stay cached.
RUN python3 -m pip install thirdai[neural_db_v2]
RUN python3 -m pip install -U langchain-community

COPY ./. /app/

CMD ["uvicorn","main:app","--reload","--host","0.0.0.0","--port","80"]