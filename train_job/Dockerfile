# Use the official FastAPI image with Python 3.9
FROM tiangolo/uvicorn-gunicorn-fastapi:python3.9

# Set the working directory in the container to /app
WORKDIR /app

# Copy the requirements.txt file into the container
COPY requirements.txt /app/

# Install dependencies from requirements.txt
RUN python3 -m pip install --no-cache-dir -r requirements.txt
RUN python3 -m pip install pdftitle

COPY ./thirdai-0.9.14+9d0ffa8-cp39-cp39-manylinux_2_17_x86_64.manylinux2014_x86_64.whl /app/
RUN python3 -m pip install thirdai-0.9.14+9d0ffa8-cp39-cp39-manylinux_2_17_x86_64.manylinux2014_x86_64.whl --force-reinstall

# Copy the current directory contents into the container at /app
COPY . /app

# Set the default command to run the FastAPI application
CMD ["python3", "run.py"]