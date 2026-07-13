# WSMC observability

Set `OTEL_METRICS_ENABLED=true` and configure Gate's normal OTLP or Prometheus exporter. The WSMC
listener publishes low-cardinality transport metrics. Client IP addresses and literal virtual hosts
are intentionally not used as labels.

| Metric | Purpose |
| --- | --- |
| `gate.wsmc.active_sessions` | Current accepted sessions by mode |
| `gate.wsmc.requests` | Accepted and rejected handshakes by result and mode |
| `gate.wsmc.bytes` | Stream bytes by direction and mode |
| `gate.wsmc.frames` | Binary messages emitted by Gate |
| `gate.wsmc.write.duration` | Time spent writing one coalesced batch |
| `gate.wsmc.batch.bytes` | Coalesced stream batch size |
| `gate.wsmc.pending.bytes` | Pending bytes after a completed write |
| `gate.wsmc.target_frame.bytes` | Adaptive outgoing message payload target |
| `gate.wsmc.handshake.duration` | HTTP upgrade duration by result and mode |
| `gate.wsmc.session.duration` | Accepted session lifetime by mode |
| `gate.wsmc.events` | Idle, backpressure, close-flush and write failures |

The compatibility test follows `rikka0w0/wsmc` commit
`ba665afdf3bd3d1feb90431dd0d53cd077b98f61`: RFC 6455 V13, binary messages, a custom HTTP Host,
and a default maximum frame payload of 65536 bytes. WSMC treats message payloads as a continuous
Minecraft TCP byte stream, so Gate may safely coalesce and re-slice messages without adding an
application-level acknowledgement or retransmission protocol.
