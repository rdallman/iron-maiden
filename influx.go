package main

import (
	"fmt"
	"log"
	"net/url"
	"time"

	influxClient "github.com/influxdb/influxdb/client"
	"github.com/rcrowley/go-metrics"
)

const (
	msi = int64(time.Millisecond)
	ms  = float64(time.Millisecond)
)

// Most the code copied from here
// https://github.com/rcrowley/go-metrics/blob/master/influxdb/influxdb.go
type Config struct {
	Host     string `json:host`
	Database string `json:database`
	Username string `json:username`
	Password string `json:password`
}

func Influxdb(r metrics.Registry, d time.Duration, config *Config) {
	uri, err := url.Parse(fmt.Sprintf("http://%s", config.Host))

	client, err := influxClient.NewClient(influxClient.Config{
		URL:      *uri,
		Username: config.Username,
		Password: config.Password,
	})

	if err != nil {
		log.Println(err)
		return
	}

	for _ = range time.Tick(d) {
		if err := send(r, client, config.Database); err != nil {
			log.Println(err)
		}
	}
}

func send(r metrics.Registry, client *influxClient.Client, db string) error {
	points := make([]influxClient.Point, 0)

	r.Each(func(name string, i interface{}) {
		now := time.Now()
		switch metric := i.(type) {
		case metrics.Timer:
			h := metric.Snapshot()
			ps := h.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
			points = append(points, influxClient.Point{
				Measurement: fmt.Sprintf("%s", name),
				Time:        now,
				Fields: map[string]interface{}{"count": h.Count(),
					"min":            h.Min(),
					"max":            h.Max(),
					"mean":           h.Mean(),
					"std-dev":        h.StdDev(),
					"50-percentile":  ps[0],
					"75-percentile":  ps[1],
					"95-percentile":  ps[2],
					"99-percentile":  ps[3],
					"999-percentile": ps[4],
					"one-minute":     h.Rate1(),
					"five-minute":    h.Rate5(),
					"fifteen-minute": h.Rate15(),
					"mean-rate":      h.RateMean(),
				},
			})
		}
	})

	bp := influxClient.BatchPoints{
		Points:          points,
		Database:        db,
		RetentionPolicy: "default",
	}
	if _, err := client.Write(bp); err != nil {
		log.Println(err)
	}
	return nil
}
