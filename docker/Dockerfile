FROM golang:1.13 AS builder

RUN mkdir -p /go/src/github.com/builder/cloud-consumer

ADD cloud-consumer/main.go /go/src/github.com//builder/cloud-consumer
ADD cloud-consumer/go.mod /go/src/github.com//builder/cloud-consumer
ADD cloud-consumer/go.sum /go/src/github.com//builder/cloud-consumer

WORKDIR /go/src/github.com/builder/cloud-consumer

RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /go/bin/consumer

FROM alpine:3.11.2

RUN apk add --no-cache ca-certificates openssl

COPY --from=builder /go/bin/consumer /go/bin/consumer

EXPOSE 8081

ENTRYPOINT [ "/go/bin/consumer" ]
