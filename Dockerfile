FROM golang:1.24 AS base

WORKDIR /build

COPY ./cmd ./cmd
COPY ./config ./config
COPY ./log ./log
COPY ./rest ./rest
COPY ./services ./services
COPY go.mod ./
COPY go.sum ./
RUN go mod download

RUN go build -o ./notification ./cmd/main.go

FROM alpine:3.22

RUN apk add --no-cache libstdc++ gcompat libgomp
RUN apk add --update busybox>1.3.1-r0
RUN apk add --update openssl>3.1.4-r1

COPY --from=base /build/notification       /app/notification
COPY --from=base  /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app

# Command to run
ENTRYPOINT ["/app/notification"]
