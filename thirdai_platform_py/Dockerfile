FROM tiangolo/uvicorn-gunicorn-fastapi:python3.11

WORKDIR /app

# Install PostgreSQL client version 17.x
RUN apt-get update && \
    apt-get install -y wget gnupg2 lsb-release && \
    echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list && \
    wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - && \
    apt-get update && \
    apt-get install -y postgresql-client-17


# Define build argument to toggle between local and PyPI versions of thirdai
ARG use_local_wheel=true

# Copy the local wheel only if using local setup
COPY files/thirdai-0.9.27-cp311-cp311-manylinux_2_17_aarch64.manylinux2014_aarch64.whl /app/thirdai-0.9.27-cp311-cp311-manylinux_2_17_aarch64.manylinux2014_aarch64.whl

# TODO(pratik): Add actions to build arm wheels and push to PyPI
# Install thirdai package based on the argument
RUN if [ "$use_local_wheel" = "true" ]; then \
    python3 -m pip install /app/thirdai-0.9.27-cp311-cp311-manylinux_2_17_aarch64.manylinux2014_aarch64.whl[neural_db_v2]; \
    else \
    python3 -m pip install "thirdai[neural_db_v2]>=0.9.21"; \
    fi

# Install dependencies before copying the directory so python dependencies are 
# cached even when there are changes in the directory.
COPY requirements.txt /app/requirements.txt
RUN pip install -r /app/requirements.txt

# Ensure latest version of langchain-community is installed
RUN python3 -m pip install -U langchain-community

COPY . /app

# Disable outbound telemetry for Unstructured and HuggingFace
ENV SCARF_NO_ANALYTICS=true
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
