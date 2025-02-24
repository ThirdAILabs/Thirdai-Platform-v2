FROM tiangolo/uvicorn-gunicorn-fastapi:python3.11

WORKDIR /app

# Install PostgreSQL client version 17.x
RUN apt-get update && \
    apt-get install -y wget gnupg2 lsb-release && \
    echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list && \
    wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - && \
    apt-get update && \
    apt-get install -y postgresql-client-17

# Install dependencies before copying the directory so python dependencies are 
# cached even when there are changes in the directory.
COPY requirements.txt /app/requirements.txt
RUN pip install -r /app/requirements.txt

# Copy the local wheel from ../local_setup/files/ to the container
COPY files/thirdai-0.9.27-cp311-cp311-manylinux_2_17_aarch64.manylinux2014_aarch64.whl /app/thirdai-0.9.27-cp311-cp311-manylinux_2_17_aarch64.manylinux2014_aarch64.whl


# Install thirdai package separately since this is the dependency that changes
# the most often. This allows the other packages to stay cached.
RUN python3 -m pip install /app/thirdai-0.9.27-cp311-cp311-manylinux_2_17_aarch64.manylinux2014_aarch64.whl[neural_db_v2]
RUN python3 -m pip install -U langchain-community

COPY . /app

# Disable outbound telemetry for the Unstructured library
ENV SCARF_NO_ANALYTICS=true

# Disable outbound telemetry for the HuggingFace library
ENV HF_HUB_DISABLE_TELEMETRY=1

# Download nltk models
RUN python3 -m nltk.downloader -d /usr/share/nltk_data stopwords punkt_tab

# Use build args instead of env variables to pass in these variables at build time
ARG tag
ARG docker_registry
ARG docker_username
ARG docker_password
ARG export_image_names_command

ENV TAG $tag
ENV DOCKER_REGISTRY $docker_registry
ENV DOCKER_USERNAME $docker_username
ENV DOCKER_PASSWORD $docker_password
ENV PLATFORM docker
ENV EXPORT_IMAGE_NAMES_COMMAND $export_image_names_command

RUN chmod +x ./entrypoint.sh

CMD ["./entrypoint.sh"]