FROM alpine:3.4

RUN apk add --update ca-certificates

COPY prom-sidekiq-exporter /usr/local/bin/prom-sidekiq-exporter

ENTRYPOINT ["/usr/local/bin/prom-sidekiq-exporter"]
