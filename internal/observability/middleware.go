package observability

import (
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func WrapHTTPHandler(handler http.Handler, metrics *Metrics) http.Handler {
	// Wrap with OpenTelemetry instrumentation
	wrappedHandler := otelhttp.NewHandler(handler, "http-request")

	// Wrap with custom metrics middleware
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.HTTPRequestsTotal.Inc()
		metrics.HTTPRequestsInFlight.Inc()
		defer metrics.HTTPRequestsInFlight.Dec()

		start := time.Now()
		wrappedHandler.ServeHTTP(w, r)
		duration := time.Since(start).Seconds()

		metrics.HTTPRequestDuration.Observe(duration)
	})
}

// RecordDatabaseQueryDuration records the duration of database queries
func RecordDatabaseQueryDuration(metrics *Metrics, duration time.Duration) {
	metrics.DatabaseQueryDuration.Observe(duration.Seconds())
}

// RecordRabbitMQPublish records RabbitMQ publish operations
func RecordRabbitMQPublish(metrics *Metrics, success bool) {
	metrics.RabbitMQPublishTotal.Inc()
	if !success {
		metrics.RabbitMQPublishErrors.Inc()
	}
}
