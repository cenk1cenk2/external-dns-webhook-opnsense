package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

const namespace = "opnsense_webhook"

type Metrics struct {
	Registry *prometheus.Registry

	HttpRequests        *prometheus.CounterVec
	HttpRequestDuration *prometheus.HistogramVec

	ClientRequests        *prometheus.CounterVec
	ClientRequestDuration *prometheus.HistogramVec
	ClientRetries         prometheus.Counter

	UnboundReconfigures *prometheus.CounterVec
}

func New() *Metrics {
	m := &Metrics{
		Registry: prometheus.NewRegistry(),

		HttpRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests handled by the webhook server.",
		}, []string{"method", "route", "code"}),
		HttpRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "Duration of HTTP requests handled by the webhook server.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "route"}),

		ClientRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "client_requests_total",
			Help:      "Total number of OPNsense API calls, retries included in a single call. The code is \"error\" when no response was received.",
		}, []string{"method", "endpoint", "code"}),
		ClientRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "client_request_duration_seconds",
			Help:      "Duration of OPNsense API calls including retries and backoff.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "endpoint"}),
		ClientRetries: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "client_retries_total",
			Help:      "Total number of OPNsense API request retries.",
		}),

		UnboundReconfigures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "unbound_reconfigures_total",
			Help:      "Total number of Unbound service reconfigure calls by result.",
		}, []string{"result"}),
	}

	m.Registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		m.HttpRequests,
		m.HttpRequestDuration,
		m.ClientRequests,
		m.ClientRequestDuration,
		m.ClientRetries,
		m.UnboundReconfigures,
	)

	return m
}
