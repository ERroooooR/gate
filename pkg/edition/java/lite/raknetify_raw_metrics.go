package lite

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var rawRaknetifyMetrics = newRawRaknetifyMetrics()

type rawRaknetifyMetricsRecorder struct {
	activeSessions metric.Int64UpDownCounter
	droppedPackets metric.Int64Counter
	sessionEvents  metric.Int64Counter
	writeFailures  metric.Int64Counter
}

func newRawRaknetifyMetrics() rawRaknetifyMetricsRecorder {
	meter := otel.Meter("java/lite")
	activeSessions, _ := meter.Int64UpDownCounter(
		"gate.raknetify.raw.active_sessions",
		metric.WithDescription("The current number of raw Raknetify passthrough sessions"),
		metric.WithUnit("1"),
	)
	droppedPackets, _ := meter.Int64Counter(
		"gate.raknetify.raw.dropped_packets",
		metric.WithDescription("The number of raw Raknetify passthrough packets dropped by Gate"),
		metric.WithUnit("1"),
	)
	sessionEvents, _ := meter.Int64Counter(
		"gate.raknetify.raw.session_events",
		metric.WithDescription("The number of raw Raknetify passthrough session lifecycle events"),
		metric.WithUnit("1"),
	)
	writeFailures, _ := meter.Int64Counter(
		"gate.raknetify.raw.write_failures",
		metric.WithDescription("The number of raw Raknetify passthrough UDP write failures"),
		metric.WithUnit("1"),
	)
	return rawRaknetifyMetricsRecorder{
		activeSessions: activeSessions,
		droppedPackets: droppedPackets,
		sessionEvents:  sessionEvents,
		writeFailures:  writeFailures,
	}
}

func (m rawRaknetifyMetricsRecorder) addActiveSessions(delta int64) {
	if m.activeSessions == nil {
		return
	}
	m.activeSessions.Add(context.Background(), delta)
}

func (m rawRaknetifyMetricsRecorder) recordDroppedPacket(direction, reason string) {
	if m.droppedPackets == nil {
		return
	}
	m.droppedPackets.Add(
		context.Background(),
		1,
		metric.WithAttributes(
			attribute.String("direction", direction),
			attribute.String("reason", reason),
		),
	)
}

func (m rawRaknetifyMetricsRecorder) recordSessionEvent(event, reason string) {
	if m.sessionEvents == nil {
		return
	}
	attrs := []attribute.KeyValue{attribute.String("event", event)}
	if reason != "" {
		attrs = append(attrs, attribute.String("reason", reason))
	}
	m.sessionEvents.Add(context.Background(), 1, metric.WithAttributes(attrs...))
}

func (m rawRaknetifyMetricsRecorder) recordWriteFailure(direction, reason string) {
	if m.writeFailures == nil {
		return
	}
	m.writeFailures.Add(
		context.Background(),
		1,
		metric.WithAttributes(
			attribute.String("direction", direction),
			attribute.String("reason", reason),
		),
	)
}
