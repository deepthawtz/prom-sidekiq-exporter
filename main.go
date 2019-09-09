package main

import (
	"io/ioutil"
	"net/http"

	yaml "gopkg.in/yaml.v2"

	"github.com/deepthawtz/prom-sidekiq-exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func main() {
	y, err := ioutil.ReadFile("./config.yml")
	if err != nil {
		logrus.Fatal(err)
	}

	c := &exporter.Config{}
	if err := yaml.Unmarshal(y, c); err != nil {
		logrus.Fatal(err)
	}

	exp, err := exporter.NewExporter(c)
	if err != nil {
		logrus.Fatal(err)
	}

	prometheus.MustRegister(exp)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>Prometheus Sidekiq Exporter</title></head>
		<body>
		<h1>Prometheus Sidekiq Exporter</h1>
		<p><a href="/metrics">Metrics</a></p>
		</body></html>`))
	})

	logrus.Info("Listening on port 9090")
	logrus.Fatal(http.ListenAndServe(":9090", nil))
}
