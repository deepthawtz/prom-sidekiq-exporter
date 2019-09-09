FROM alpine:3.4

RUN apk add --update ca-certificates

COPY prom-sidekiq-exporter /app/prom-sidekiq-exporter

WORKDIR /app
ENTRYPOINT ["prom-sidekiq-exporter"]
