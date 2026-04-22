FROM golang:1.26.1-alpine AS builder

WORKDIR /src

RUN apk --update add ca-certificates
RUN update-ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-s -w' -o bin/remove-default-vpc .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /src/bin/remove-default-vpc /remove-default-vpc
ENTRYPOINT ["/remove-default-vpc"]