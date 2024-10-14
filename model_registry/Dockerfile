FROM golang:1.23

WORKDIR /app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -v

ENV registry_db="data/registry.db"
ENV registry_storage="data/storage"
ENV registry_port="8080"
ENV admin_email=""
ENV admin_password=""

CMD ["./model_registry"]