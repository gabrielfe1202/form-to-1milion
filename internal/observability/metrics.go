package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	HTTPRequestsTotal     prometheus.Counter
	HTTPRequestDuration   prometheus.Histogram
	HTTPRequestsInFlight  prometheus.Gauge
	UserCreatedTotal      prometheus.Counter
	UserListTotal         prometheus.Counter
	UserCountTotal        prometheus.Counter
	DatabaseQueryDuration prometheus.Histogram
	RabbitMQPublishTotal  prometheus.Counter
	RabbitMQPublishErrors prometheus.Counter
}

func NewMetrics() *Metrics {
	return &Metrics{
		HTTPRequestsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		}),
		HTTPRequestDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		HTTPRequestsInFlight: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests in flight",
		}),
		UserCreatedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "user_created_total",
			Help: "Total number of users created",
		}),
		UserListTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "user_list_total",
			Help: "Total number of user list requests",
		}),
		UserCountTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "user_count_total",
			Help: "Total number of user count requests",
		}),
		DatabaseQueryDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		RabbitMQPublishTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "rabbitmq_publish_total",
			Help: "Total number of RabbitMQ publishes",
		}),
		RabbitMQPublishErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "rabbitmq_publish_errors_total",
			Help: "Total number of RabbitMQ publish errors",
		}),
	}
}
