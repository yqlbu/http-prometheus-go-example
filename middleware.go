// The following code shows the example code for the middleware:
// Source: https://robert-scherbarth.medium.com/measure-request-duration-with-prometheus-and-golang-adc6f4ca05fe

package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// statusRecorder to record the status code from the ResponseWriter
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(statusCode int) {
	rec.statusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

func measureResponseDuration(next http.Handler) http.Handler {
	buckets := []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

	responseTimeHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "namespace",
		Name:      "http_server_request_duration_seconds",
		Help:      "Histogram of response time for handler in seconds",
		Buckets:   buckets,
	}, []string{"route", "method", "status_code"})

	prometheus.MustRegister(responseTimeHistogram)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := statusRecorder{w, 200}

		next.ServeHTTP(&rec, r)

		duration := time.Since(start)
		statusCode := strconv.Itoa(rec.statusCode)
		route := getRoutePattern(r)
		responseTimeHistogram.WithLabelValues(route, r.Method, statusCode).Observe(duration.Seconds())
	})
}

// getRoutePattern returns the route pattern from the chi context there are 3 conditions
// a) static routes "/example" => "/example"
// b) dynamic routes "/example/:id" => "/example/{id}"
// c) if nothing matches the output is undefined
func getRoutePattern(r *http.Request) string {
	reqContext := chi.RouteContext(r.Context())
	if pattern := reqContext.RoutePattern(); pattern != "" {
		return pattern
	}

	return "undefined"
}
