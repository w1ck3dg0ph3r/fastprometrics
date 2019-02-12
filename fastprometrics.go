/*
Package fastprometrics provides fasthttp prometheus metrics middleware

Example

	router := fasthttprouter.New()
	handler := router.Handler
	handler = metrics.Add(handler,
		metrics.WithPath("/metrics"),
		metrics.WithSubsystem("http"),
	)
	fasthttp.ListenAndServe("127.0.0.1:8080", handler)

*/
package fastprometrics

import (
	"bytes"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// WithPath configures metrics path
// Default is "/metrics"
func WithPath(path string) Options {
	return func(o *options) {
		o.metricsPath = []byte(path)
	}
}

// WithSubsystem configures subsystem name
func WithSubsystem(subsystem string) Options {
	return func(o *options) {
		o.metricsSubsystem = subsystem
	}
}

// WithLatencyBuckets configures latency buckets in milliseconds
// Default buckets are 1, 3, 5, 10, 15, 25, 50, 100, 200, 300, 500, 1000, 2000, 5000, 10000 milliseconds
func WithLatencyBuckets(buckets []float64) Options {
	return func(o *options) {
		o.latencyBuckets = buckets
	}
}

// Add wraps provided handler adding metrics endpoint and metrics gathering
func Add(h fasthttp.RequestHandler, os ...Options) fasthttp.RequestHandler {
	// Default options
	opts := options{
		metricsPath:       []byte("/metrics"),
		metricsSubsystem:  "http",
		prometheusHandler: fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler()),
		latencyBuckets:    []float64{1, 3, 5, 10, 15, 25, 50, 100, 200, 300, 500, 1000, 2000, 5000, 10000},
	}
	// Apply specified options
	for _, o := range os {
		o(&opts)
	}

	opts.registerMetrics()

	return func(c *fasthttp.RequestCtx) {
		// Handle metrics endpoint
		if bytes.Equal(c.RequestURI(), opts.metricsPath) {
			opts.prometheusHandler(c)
			return
		}

		start := time.Now()
		h(c)
		status := strconv.Itoa(c.Response.StatusCode())
		latency := float64(time.Since(start) / time.Millisecond)
		opts.requestCounter.WithLabelValues(status).Inc()
		opts.requestLatency.WithLabelValues(status).Observe(float64(latency))
	}
}

type options struct {
	metricsPath      []byte
	metricsSubsystem string

	prometheusHandler fasthttp.RequestHandler

	requestCounter *prometheus.CounterVec
	requestLatency *prometheus.HistogramVec
	reuestSize     prometheus.Summary
	responseSize   prometheus.Summary

	latencyBuckets []float64
}

func (o *options) registerMetrics() {
	o.requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: o.metricsSubsystem,
			Name:      "requests_total",
			Help:      "The amount of HTTP requests processed.",
		},
		[]string{"code"},
	)

	o.requestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: o.metricsSubsystem,
			Name:      "request_duration_ms",
			Help:      "The HTTP request duration in milliseconds.",
			Buckets:   o.latencyBuckets,
		},
		[]string{"code"},
	)

	prometheus.MustRegister(o.requestCounter, o.requestLatency)
}

// Options is a configuration function
type Options func(o *options)
