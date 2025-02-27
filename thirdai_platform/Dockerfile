FROM golang:1.23 AS build-stage
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN mkdir -p bin
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/model_bazaar -v ./cmd/model_bazaar/main.go 

FROM gcr.io/distroless/base-debian11 AS build-release-stage
WORKDIR /app

COPY --from=build-stage /app/bin/* .

ARG TAG
ARG DOCKER_REGISTRY
ARG DOCKER_USERNAME
ARG DOCKER_PASSWORD
ARG JOBS_IMAGE_NAME
ARG FRONTEND_IMAGE_NAME

ENV TAG=$TAG
ENV DOCKER_REGISTRY=$DOCKER_REGISTRY
ENV DOCKER_USERNAME=$DOCKER_USERNAME
ENV DOCKER_PASSWORD=$DOCKER_PASSWORD
ENV JOBS_IMAGE_NAME=$JOBS_IMAGE_NAME
ENV FRONTEND_IMAGE_NAME=$FRONTEND_IMAGE_NAME

CMD ["./model_bazaar", "-port=8000"]