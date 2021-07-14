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

# # Build the producer
# WORKDIR /producer
# COPY go.mod .
# COPY producer.go .
# RUN go build -o producer .
# RUN rm producer.go && rm go.*

# CMD ["./producer"]
# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from golang:1.12-alpine base image
FROM golang:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependancies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o producer .

# Expose port 8080 to the outside world
EXPOSE 8080

# Run the executable
CMD ["./producer"]

