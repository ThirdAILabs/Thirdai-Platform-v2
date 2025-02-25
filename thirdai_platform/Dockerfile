FROM golang:1.23 AS build-stage
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN mkdir -p bin
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/model_bazaar -v ./cmd/model_bazaar/main.go 

# Per https://docs.docker.com/guides/golang/build-images/#multi-stage-builds
FROM gcr.io/distroless/base-debian11 AS build-release-stage
WORKDIR /app

COPY --from=build-stage /app/bin/* .


ENV TAG="latest"
ENV DOCKER_REGISTRY="thirdaiplatform.azurecr.io"
ENV DOCKER_USERNAME="thirdaiplatform-pull-local-dev"
ENV DOCKER_PASSWORD="MBe9v3dhcRf1ZiroZEkBTz1foVJPHa9hPG6FcLzYWu+ACRDGHVcp"
ENV JOBS_IMAGE_NAME="thirdai_platform_jobs_local-dev"
ENV FRONTEND_IMAGE_NAME=""

CMD ["./model_bazaar", "-port=80"]