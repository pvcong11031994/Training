FROM golang:1.14.2-alpine as builder
RUN apk add alpine-sdk
WORKDIR /go/app
COPY main.go /go/app
COPY go.mod /go/app
COPY go.sum /go/app
RUN go mod download
RUN GOOS=linux GOARCH=amd64 go build -o main -tags musl

FROM alpine:latest as runner
WORKDIR /root/
COPY --from=builder /go/app/main .
EXPOSE 8081
ENTRYPOINT /root/main