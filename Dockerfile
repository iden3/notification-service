##
## Build did driver
##
FROM golang:1.19.7-alpine as base

WORKDIR /build

COPY ./cmd ./cmd
COPY ./config ./config
COPY ./log ./log
COPY ./rest ./rest
COPY ./services ./services
COPY go.mod ./
COPY go.sum ./
RUN go mod download

RUN CGO_ENABLED=0 go build -o ./notification ./cmd/main.go

# Build an driver image
FROM scratch

COPY --from=base /build/notification       /app/notification
COPY --from=base  /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app

# Command to run
ENTRYPOINT ["/app/notification"]
