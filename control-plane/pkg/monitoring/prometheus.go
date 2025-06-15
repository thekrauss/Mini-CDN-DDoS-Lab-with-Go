package monitoring

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	GRPCRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Nombre total de requêtes gRPC traitées.",
		},
		[]string{"method", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Durée des requêtes gRPC",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	PingCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "node_ping_total",
			Help: "Nombre total de Ping() reçus",
		},
	)
)

func Init() {
	prometheus.MustRegister(GRPCRequests)
	prometheus.MustRegister(RequestDuration)
	prometheus.MustRegister(PingCounter)
}

func Handler() http.Handler {
	return promhttp.Handler()
}
