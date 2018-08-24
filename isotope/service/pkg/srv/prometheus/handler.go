// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this currentFile except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prometheus

import (
	"net/http"
	"strconv"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	durationBuckets = []float64{
		0.007, 0.008, 0.009, 0.01, 0.011, 0.012, 0.014, 0.016, 0.018, 0.02, 0.025,
		0.03, 0.035, 0.04, 0.045, 0.05, 0.06, 0.07, 0.08, 0.09, 0.1, 0.12, 0.14,
		0.16, 0.18, 0.2, 0.25, 0.3, 0.35, 0.4, 0.45, 0.5}
	sizeBuckets = []float64{
		// 1, 10, 100, 1,000, ..., 1,000,000,000
		1e+00, 1e+01, 1e+02, 1e+03, 1e+04, 1e+05, 1e+06, 1e+07, 1e+08, 1e+09}

	serviceIncomingRequestsTotal = prom.NewCounter(
		prom.CounterOpts{
			Name: "service_incoming_requests_total",
			Help: "Number of requests sent to this service.",
		})

	serviceOutgoingRequestsTotal = prom.NewCounterVec(
		prom.CounterOpts{
			Name: "service_outgoing_requests_total",
			Help: "Number of requests sent from this service.",
		}, []string{"destination_service"})

	serviceOutgoingRequestSize = prom.NewHistogramVec(
		prom.HistogramOpts{
			Name:    "service_outgoing_request_size",
			Help:    "Size in bytes of requests sent from this service.",
			Buckets: sizeBuckets,
		}, []string{"destination_service"})

	serviceRequestDurationSeconds = prom.NewHistogramVec(
		prom.HistogramOpts{
			Name:    "service_request_duration_seconds",
			Help:    "Duration in seconds it took to serve requests to this service.",
			Buckets: durationBuckets,
		}, []string{"code"})

	serviceResponseSize = prom.NewHistogramVec(
		prom.HistogramOpts{
			Name:    "service_response_size",
			Help:    "Size in bytes of responses sent from this service.",
			Buckets: sizeBuckets,
		}, []string{"code"})
)

// Handler returns an http.Handler which should be attached to a "/metrics"
// endpoint for Prometheus to ingest.
func Handler() http.Handler {
	prom.MustRegister(serviceIncomingRequestsTotal)

	prom.MustRegister(serviceOutgoingRequestsTotal)
	prom.MustRegister(serviceOutgoingRequestSize)

	prom.MustRegister(serviceRequestDurationSeconds)
	prom.MustRegister(serviceResponseSize)

	return promhttp.Handler()
}

// RecordRequestReceived increments the Prometheus counter for incoming
// requests.
func RecordRequestReceived() {
	serviceIncomingRequestsTotal.Inc()
}

// RecordRequestSent increments the Prometheus counter for outgoing requests
// and records an outgoing request size.
func RecordRequestSent(destinationService string, size uint64) {
	serviceOutgoingRequestsTotal.WithLabelValues(destinationService).Inc()
	serviceOutgoingRequestSize.WithLabelValues(destinationService).Observe(
		float64(size))
}

// RecordResponseSent observes the time-to-response duration and size for the
// HTTP status code.
func RecordResponseSent(duration time.Duration, size int, code int) {
	strCode := strconv.Itoa(code)
	serviceRequestDurationSeconds.WithLabelValues(strCode).Observe(
		duration.Seconds())
	serviceResponseSize.WithLabelValues(strCode).Observe(float64(size))
}
