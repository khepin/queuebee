FROM golang:bookworm AS builder

WORKDIR /var/www

COPY . /var/www

RUN CGO_ENABLED=1 go build -o queuebee

FROM debian:bookworm

COPY --from=builder /var/www/queuebee /usr/local/bin/queuebee

ENTRYPOINT [ "queuebee" ]
