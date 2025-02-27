FROM golang:1.23 AS build-stage
WORKDIR /app

RUN apt-get update && apt-get install -y \
    libssl-dev \
    libssl3

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN mkdir -p bin
RUN CGO_ENABLED=1 GOOS=linux go build -o bin/model_bazaar -v ./cmd/model_bazaar/main.go 

RUN ldd bin/model_bazaar

# Per https://docs.docker.com/guides/golang/build-images/#multi-stage-builds
FROM gcr.io/distroless/base-debian11 AS build-release-stage
WORKDIR /app

COPY --from=build-stage /app/bin/* .

# Copy SSL libraries from their actual location
COPY --from=build-stage /usr/lib/x86_64-linux-gnu/libssl.so* /lib/x86_64-linux-gnu/
COPY --from=build-stage /usr/lib/x86_64-linux-gnu/libcrypto.so* /lib/x86_64-linux-gnu/

# Copy other libraries
COPY --from=build-stage /usr/lib/x86_64-linux-gnu/libstdc++.so* /lib/x86_64-linux-gnu/
COPY --from=build-stage /usr/lib/x86_64-linux-gnu/libm.so* /lib/x86_64-linux-gnu/
COPY --from=build-stage /usr/lib/x86_64-linux-gnu/libgomp.so* /lib/x86_64-linux-gnu/
COPY --from=build-stage /usr/lib/x86_64-linux-gnu/libgcc_s.so* /lib/x86_64-linux-gnu/
COPY --from=build-stage /usr/lib/x86_64-linux-gnu/libc.so* /lib/x86_64-linux-gnu/
COPY --from=build-stage /usr/lib64/ld-linux-x86-64.so* /lib64/


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