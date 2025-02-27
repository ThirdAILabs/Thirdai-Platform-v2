FROM golang:1.23 AS build-stage
WORKDIR /app

RUN apt-get update && apt-get install -y libssl-dev

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN mkdir -p bin
RUN CGO_ENABLED=1 GOOS=linux go build -o bin/model_bazaar -v ./cmd/model_bazaar/main.go 

# Per https://docs.docker.com/guides/golang/build-images/#multi-stage-builds
FROM ubuntu:22.04 AS build-release-stage
WORKDIR /app

COPY --from=build-stage /app/bin/* .

RUN apt-get update && apt-get install -y \
    libssl3 \
    libstdc++6 \
    libgomp1 \
    && rm -rf /var/lib/apt/lists/*

ARG tag
ARG docker_registry
ARG docker_username
ARG docker_password
ARG thirdai_platform_jobs
ARG thirdai_platform_frontend

ENV TAG=$tag
ENV DOCKER_REGISTRY=$docker_registry
ENV DOCKER_USERNAME=$docker_username
ENV DOCKER_PASSWORD=$docker_password
ENV JOBS_IMAGE_NAME=$thirdai_platform_jobs
ENV FRONTEND_IMAGE_NAME=$thirdai_platform_frontend

CMD ["./model_bazaar", "-port=80"]