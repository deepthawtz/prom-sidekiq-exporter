package exporter

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	redis "gopkg.in/redis.v3"
)

var mux sync.Mutex

// Config holds list of apps to collect Sidekiq redis metrics for
type Config struct {
	Collectors []struct {
		App      string `yaml:"app"`
		Env      string `yaml:"env"`
		RedisURI string `yaml:"redis-uri"`
	} `yaml:"collectors"`
}

// Collector represents an individual Sidekiq apps info
type Collector struct {
	App           string
	Env           string
	RedisClient   *redis.Client
	QueueCounters map[string]*prometheus.Desc
}

// Exporter collects and exports prometheus metrics
type Exporter struct {
	Collectors []*Collector
}

// NewExporter returns an initialized Exporter.
func NewExporter(config *Config) (*Exporter, error) {
	exp := &Exporter{}
	var wg sync.WaitGroup
	for _, c := range config.Collectors {
		wg.Add(1)
		go func(app, env, uri string) error {
			defer wg.Done()
			if err := validateRedisURI(uri); err != nil {
				return err
			}
			// we've validated for properly formatted redis URI
			u, _ := url.Parse(uri)
			p := strings.Split(u.Path, "/")
			db, _ := strconv.Atoi(p[len(p)-1])
			client := redis.NewClient(&redis.Options{
				Addr: u.Host,
				DB:   int64(db),
			})

			coll := &Collector{
				App:         app,
				Env:         env,
				RedisClient: client,
			}

			cmd := client.Keys("queue:*")
			if err := cmd.Err(); err != nil {
				logrus.Error(err)
			}
			queues := []string{"dead"}
			for _, v := range cmd.Val() {
				n := strings.Split(v, ":")
				if len(n) == 2 {
					queues = append(queues, v)
				}
			}
			coll.QueueCounters = map[string]*prometheus.Desc{}
			for _, q := range queues {
				coll.QueueCounters[q] = prometheus.NewDesc(
					"sidekiq_queue_count",
					"length of important sidekiq queues",
					[]string{"app", "env", "queue"}, nil,
				)
			}

			mux.Lock()
			exp.Collectors = append(exp.Collectors, coll)
			mux.Unlock()

			return nil
		}(c.App, c.Env, c.RedisURI)
	}
	wg.Wait()

	return exp, nil
}

// Describe satifies prometheus Collector interface
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, c := range e.Collectors {
		for _, q := range c.QueueCounters {
			ch <- q
		}
	}
}

// Collect satifies prometheus Collector interface
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	for _, c := range e.Collectors {
		for k, q := range c.QueueCounters {
			parts := strings.Split(k, ":")
			var queue string
			if len(parts) == 2 {
				queue = parts[1]
			} else {
				queue = parts[0]
			}
			var val float64
			if queue == "dead" {
				cmd := c.RedisClient.ZCard(k)
				if err := cmd.Err(); err != nil {
					logrus.Error(err)
				}
				val = float64(cmd.Val())
			} else {
				cmd := c.RedisClient.LLen(k)
				if err := cmd.Err(); err != nil {
					logrus.Error(err)
				}
				val = float64(cmd.Val())
			}
			ch <- prometheus.MustNewConstMetric(q, prometheus.GaugeValue, val, c.App, c.Env, queue)
		}
	}
}

func validateRedisURI(uri string) error {
	u, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("redis_url (%s) failed to parse: %s", uri, err)
	}
	p := strings.Split(u.Path, "/")
	if len(p) != 2 {
		return fmt.Errorf("redis_url (%s) must be in redis://host:port/db format", uri)
	}
	_, err = strconv.Atoi(p[len(p)-1])
	if err != nil {
		return fmt.Errorf("redis_url (%s) must be in redis://host:port/db format: %s", uri, err)
	}

	return nil
}
