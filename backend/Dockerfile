# syntax=docker/dockerfile:1

FROM golang:1.23-bookworm

WORKDIR /var/deployment-controller/app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
COPY config-docker.yml ./config.yml

# RUN go install
RUN CGO_ENABLED=0 GOOS=linux go build -o /var/deployment-controller/build
EXPOSE 8080

CMD ["/var/deployment-controller/build"]