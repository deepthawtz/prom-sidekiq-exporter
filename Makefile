package = github.com/deepthawtz/prom-sidekiq-exporter

.PHONY: install release

build:
	goreleaser --rm-dist --skip-validate --skip-publish

release:
	goreleaser --rm-dist

install:
	cp dist/darwin_amd64/prom-sidekiq-exporter /usr/local/bin/prom-sidekiq-exporter
	chmod +x /usr/local/bin/prom-sidekiq-exporter
