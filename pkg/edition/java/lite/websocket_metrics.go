package lite

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var wsmcMetrics = newWSMCMetrics()

type wsmcMetricsRecorder struct {
	active        metric.Int64UpDownCounter
	requests      metric.Int64Counter
	bytes         metric.Int64Counter
	frames        metric.Int64Counter
	events        metric.Int64Counter
	writeDuration metric.Float64Histogram
	batchBytes    metric.Int64Histogram
	pendingBytes  metric.Int64Histogram
	targetFrame   metric.Int64Histogram
	handshakeTime metric.Float64Histogram
	sessionTime   metric.Float64Histogram
}

func newWSMCMetrics() wsmcMetricsRecorder {
	m := otel.Meter("java/lite/wsmc")
	active, _ := m.Int64UpDownCounter("gate.wsmc.active_sessions", metric.WithUnit("1"))
	requests, _ := m.Int64Counter("gate.wsmc.requests", metric.WithUnit("1"))
	bytes, _ := m.Int64Counter("gate.wsmc.bytes", metric.WithUnit("By"))
	frames, _ := m.Int64Counter("gate.wsmc.frames", metric.WithUnit("1"))
	events, _ := m.Int64Counter("gate.wsmc.events", metric.WithUnit("1"))
	writeDuration, _ := m.Float64Histogram("gate.wsmc.write.duration", metric.WithUnit("ms"))
	batchBytes, _ := m.Int64Histogram("gate.wsmc.batch.bytes", metric.WithUnit("By"))
	pendingBytes, _ := m.Int64Histogram("gate.wsmc.pending.bytes", metric.WithUnit("By"))
	targetFrame, _ := m.Int64Histogram("gate.wsmc.target_frame.bytes", metric.WithUnit("By"))
	handshakeTime, _ := m.Float64Histogram("gate.wsmc.handshake.duration", metric.WithUnit("ms"))
	sessionTime, _ := m.Float64Histogram("gate.wsmc.session.duration", metric.WithUnit("s"))
	return wsmcMetricsRecorder{active, requests, bytes, frames, events, writeDuration, batchBytes, pendingBytes, targetFrame, handshakeTime, sessionTime}
}

func modeAttr(mode string) metric.MeasurementOption {
	return metric.WithAttributes(attribute.String("mode", mode))
}

func (m wsmcMetricsRecorder) addActive(mode string, delta int64) {
	if m.active != nil {
		m.active.Add(context.Background(), delta, modeAttr(mode))
	}
}

func (m wsmcMetricsRecorder) request(result, mode string, elapsed time.Duration) {
	if m.requests != nil {
		m.requests.Add(context.Background(), 1, metric.WithAttributes(
			attribute.String("result", result), attribute.String("mode", mode)))
	}
	if m.handshakeTime != nil {
		m.handshakeTime.Record(context.Background(), float64(elapsed)/float64(time.Millisecond),
			metric.WithAttributes(attribute.String("result", result), attribute.String("mode", mode)))
	}
}

func (m wsmcMetricsRecorder) session(mode string, elapsed time.Duration) {
	if m.sessionTime != nil {
		m.sessionTime.Record(context.Background(), elapsed.Seconds(), modeAttr(mode))
	}
}

func (m wsmcMetricsRecorder) event(mode, event string) {
	if m.events != nil {
		m.events.Add(context.Background(), 1, metric.WithAttributes(
			attribute.String("mode", mode), attribute.String("event", event)))
	}
}

func (m wsmcMetricsRecorder) read(mode string, n int) {
	if n <= 0 {
		return
	}
	attrs := metric.WithAttributes(attribute.String("mode", mode), attribute.String("direction", "read"))
	if m.bytes != nil {
		m.bytes.Add(context.Background(), int64(n), attrs)
	}
}

func (m wsmcMetricsRecorder) write(mode string, frames, bytes, pending, target int, elapsed time.Duration) {
	ctx := context.Background()
	attrs := metric.WithAttributes(attribute.String("mode", mode), attribute.String("direction", "write"))
	if m.bytes != nil {
		m.bytes.Add(ctx, int64(bytes), attrs)
	}
	if m.frames != nil {
		m.frames.Add(ctx, int64(frames), modeAttr(mode))
	}
	if m.writeDuration != nil {
		m.writeDuration.Record(ctx, float64(elapsed)/float64(time.Millisecond), modeAttr(mode))
	}
	if m.batchBytes != nil {
		m.batchBytes.Record(ctx, int64(bytes), modeAttr(mode))
	}
	if m.pendingBytes != nil {
		m.pendingBytes.Record(ctx, int64(pending), modeAttr(mode))
	}
	if m.targetFrame != nil {
		m.targetFrame.Record(ctx, int64(target), modeAttr(mode))
	}
}
