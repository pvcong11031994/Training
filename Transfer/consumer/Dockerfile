# FROM fedora:latest

# LABEL maintainer="Ricardo Ferreira <riferrei@riferrei.com>"

# # Install pre-reqs
# RUN dnf install wget -y
# RUN dnf install gcc -y

# # Install librdkafka
# RUN rpm --import https://packages.confluent.io/rpm/5.4/archive.key
# COPY confluent.repo /etc/yum.repos.d
# RUN dnf install librdkafka-devel -y

# # Install Go v1.14
# RUN wget https://dl.google.com/go/go1.14.linux-amd64.tar.gz && tar -xvf go1.14.linux-amd64.tar.gz && rm go1.14.linux-amd64.tar.gz
# RUN mv go /usr/local
# ENV GOROOT=/usr/local/go
# ENV PATH="${GOROOT}/bin:${PATH}"

# # Install MongoDB.
# RUN \
#   apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 7F0CEB10 && \
#   echo 'deb http://downloads-distro.mongodb.org/repo/ubuntu-upstart dist 10gen' > /etc/apt/sources.list.d/mongodb.list && \
#   apt-get update && \
#   apt-get install -y mongodb-org && \
#   rm -rf /var/lib/apt/lists/*

# # Build the consumer
# WORKDIR /consumer
# COPY go.mod .
# COPY consumer.go .
# RUN go build -o consumer .
# RUN rm consumer.go && rm go.*

# CMD ["./consumer"]
FROM golang:1.14.2-alpine as builder
RUN apk add alpine-sdk
WORKDIR /go/app
COPY consumer.go /go/app
COPY go.mod /go/app
COPY go.sum /go/app
RUN go mod download
RUN GOOS=linux GOARCH=amd64 go build -o consumer -tags musl

FROM alpine:latest as runner
WORKDIR /root/
COPY --from=builder /go/app/consumer .
ENTRYPOINT /root/consumer

