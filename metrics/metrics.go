package metrics

import (
	"net/http"

	_ "net/http/pprof"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	RequestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "request_count",
		Help: "The total number of requests made",
	}, []string{"code"})

	RequestLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "request_latency",
		Help:    "Request latencies",
		Buckets: prometheus.ExponentialBuckets(0.02, 2, 15),
	}, []string{"code"})

	RequestInFlightCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "request_inflight_count",
		Help: "Inflight requests",
	})

	FoundURLsCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "found_urls_count",
		Help: "The total number of unique URLs found during scraping",
	})
)

func StartMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":2112", nil); err != nil {
		panic(err)
	}
}
