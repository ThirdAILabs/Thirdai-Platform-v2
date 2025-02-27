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

ARG tag
ARG docker_registry
ARG docker_username
ARG docker_password
ARG job_image_name
ARG frontend_image_name


ENV TAG $tag
ENV DOCKER_REGISTRY $docker_registry
ENV DOCKER_USERNAME $docker_username
ENV DOCKER_PASSWORD $docker_password
ENV JOBS_IMAGE_NAME=$job_image_name
ENV FRONTEND_IMAGE_NAME=$frontend_image_name

CMD ["./model_bazaar", "-port=8000"]