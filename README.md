prom-sidekiq-exporter
=====================

Prometheus exporter for current count of key Sidekiq queues. Given Sidekiq has
a 10k hard-limit on dead jobs we need to ensure we have monitoring/alerting
well before we hit that limit.

### Usage

Populate a `config.yml` file. See included example [config](./example-config.yml)

Run `go run main.go` to start the Prometheus exporter listens on localhost:9090
