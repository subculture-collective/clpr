package websocket

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// WebSocket connection metrics
	wsConnectionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "websocket_connections_total",
			Help: "Total number of WebSocket connections established",
		},
		[]string{"channel_id"},
	)

	wsConnectionsCurrent = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "websocket_connections_current",
			Help: "Current number of active WebSocket connections",
		},
		[]string{"channel_id"},
	)

	wsErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "websocket_errors_total",
			Help: "Total number of WebSocket errors",
		},
		[]string{"channel_id", "error_type"},
	)

	wsChannelsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_channels_active",
			Help: "Number of active chat channels",
		},
	)

	wsBroadcastDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "websocket_broadcast_duration_seconds",
			Help:    "Duration of message broadcast operations",
			Buckets: []float64{.0001, .0005, .001, .005, .01, .025, .05, .1},
		},
		[]string{"channel_id"},
	)
)

// RecordConnection increments connection metrics
func RecordConnection(channelID string) {
	wsConnectionsTotal.WithLabelValues(channelID).Inc()
	wsConnectionsCurrent.WithLabelValues(channelID).Inc()
}

// RecordDisconnection decrements connection metrics
func RecordDisconnection(channelID string) {
	wsConnectionsCurrent.WithLabelValues(channelID).Dec()
}

// RecordError increments error metrics
func RecordError(channelID, errorType string) {
	wsErrorsTotal.WithLabelValues(channelID, errorType).Inc()
}

// SetActiveChannels updates the active channels metric
func SetActiveChannels(count int) {
	wsChannelsActive.Set(float64(count))
}

// RecordBroadcastDuration records broadcast operation duration
func RecordBroadcastDuration(channelID string, duration float64) {
	wsBroadcastDuration.WithLabelValues(channelID).Observe(duration)
}
