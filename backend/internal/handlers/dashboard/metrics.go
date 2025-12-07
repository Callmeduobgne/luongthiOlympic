package dashboard

import (
	"sync"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// WebSocket connection metrics
	wsConnectionsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "websocket_connections_total",
		Help: "Total number of WebSocket connections established",
	})

	wsConnectionsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "websocket_connections_active",
		Help: "Number of currently active WebSocket connections",
	})

	wsAuthFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "websocket_auth_failures_total",
		Help: "Total number of WebSocket authentication failures",
	}, []string{"reason"})

	wsMessagesSentTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "websocket_messages_sent_total",
		Help: "Total number of WebSocket messages sent",
	})

	wsMessagesReceivedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "websocket_messages_received_total",
		Help: "Total number of WebSocket messages received",
	})

	wsConnectionDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "websocket_connection_duration_seconds",
		Help:    "Duration of WebSocket connections in seconds",
		Buckets: prometheus.ExponentialBuckets(1, 2, 10), // 1s to ~17min
	})
)

// ConnectionTracker tracks active WebSocket connections for monitoring
type ConnectionTracker struct {
	mu          sync.RWMutex
	connections map[string]int64 // channel -> connection count
	totalCount  int64
}

var globalConnectionTracker = &ConnectionTracker{
	connections: make(map[string]int64),
}

func (ct *ConnectionTracker) increment(channel string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.connections[channel]++
	atomic.AddInt64(&ct.totalCount, 1)
	wsConnectionsTotal.Inc()
	wsConnectionsActive.Inc()
}

func (ct *ConnectionTracker) decrement(channel string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	if ct.connections[channel] > 0 {
		ct.connections[channel]--
	}
	atomic.AddInt64(&ct.totalCount, -1)
	wsConnectionsActive.Dec()
}

func (ct *ConnectionTracker) getCount(channel string) int64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.connections[channel]
}

func (ct *ConnectionTracker) getTotalCount() int64 {
	return atomic.LoadInt64(&ct.totalCount)
}
