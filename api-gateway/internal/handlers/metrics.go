package handlers

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTPRequestsTotal counts total HTTP requests
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration measures HTTP request duration
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// BlockchainTransactionsTotal counts blockchain transactions
	BlockchainTransactionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "blockchain_transactions_total",
			Help: "Total number of blockchain transactions",
		},
		[]string{"function", "status"},
	)

	// BlockchainTransactionDuration measures blockchain transaction duration
	BlockchainTransactionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "blockchain_transaction_duration_seconds",
			Help:    "Blockchain transaction duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"function"},
	)

	// CacheHitsTotal counts cache hits
	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"key_prefix"},
	)

	// CacheMissesTotal counts cache misses
	CacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"key_prefix"},
	)

	// CircuitBreakerStateGauge tracks circuit breaker state
	CircuitBreakerStateGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=half-open, 2=open)",
		},
		[]string{"name"},
	)
)

// MetricsHandler returns Prometheus metrics handler
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

