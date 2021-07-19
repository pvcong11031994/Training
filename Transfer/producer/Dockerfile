FROM golang:1.14.2-alpine as builder
RUN apk add alpine-sdk
WORKDIR /go/app
COPY producer.go /go/app
COPY go.mod /go/app
COPY go.sum /go/app
RUN go mod download
RUN GOOS=linux GOARCH=amd64 go build -o producer -tags musl

FROM alpine:latest as runner
WORKDIR /root/
COPY --from=builder /go/app/producer .
EXPOSE 8081
ENTRYPOINT /root/producer

# Dockerizing MongoDB: 3.1 Dockerfile for building MongoDB 3.1 images
# Based on centos:centos7, installs MongoDB
# FROM centos:centos7

# MAINTAINER Jayaram Pradhan <jayaram.pradhan@uimirror.com>

# # Set up mongodb yum repo entry
# # https://www.liquidweb.com/kb/how-to-install-mongodb-on-centos-6/
# RUN echo -e "\
# [mongodb]\n\
# name=MongoDB Repository\n\
# baseurl=https://repo.mongodb.org/yum/redhat/7Server/mongodb-org/3.0/x86_64/\n\
# gpgcheck=0\n\
# enabled=1\n" >> /etc/yum.repos.d/mongodb.repo

# # Install mongodb
# RUN yum update -y && yum install -y mongodb-org

# # Set up directory requirements
# RUN mkdir -p /data/mongodb /var/log/mongodb /var/run/mongodb
# VOLUME ["/data/mongodb", "/var/log/mongodb"]

# # Expose port 27017 from the container to the host / 8081: rest api
# EXPOSE 27017

# # Start mongodb
# ENTRYPOINT ["/usr/bin/mongod"]
# CMD ["--port", "27017", "--dbpath", "/data/mongodb", "--pidfilepath", "/var/run/mongodb/mongod.pid"]