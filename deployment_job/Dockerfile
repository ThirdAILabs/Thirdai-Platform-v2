FROM tiangolo/uvicorn-gunicorn-fastapi:python3.9

WORKDIR /app

# Install dependencies before copying the directory so python dependencies are 
# cached even when there are changes in the directory.
COPY ./requirements.txt /app/requirements.txt
RUN pip install -r /app/requirements.txt

# Define a build argument with a default value
# Copy your .env file into the Docker image
COPY .env .
COPY version.txt .

# Install thirdai package separately since this is the dependency that changes
# the most often. This allows the other packages to stay cached.

RUN export THIRDAI_VERSION=$(cat "version.txt") && \
    echo "Installing thirdai version: $THIRDAI_VERSION" && \
    python3 -m pip install thirdai==$THIRDAI_VERSION

RUN python3 -m pip install thirdai[neural_db]
RUN python3 -m pip install -U langchain-community

COPY ./. /app/

CMD ["uvicorn","main:app","--reload","--host","0.0.0.0","--port","80"]