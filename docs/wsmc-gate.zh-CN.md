# Gate WSMC WebSocket 支持

Gate 的 Lite 模式可以在独立端口接受 `rikka0w0/wsmc` 客户端的 Binary WebSocket 字节流。
WSMC 的 WebSocket message 边界不属于 Minecraft 协议，因此 Gate 可以按配置重新切片。

## 监听配置

```yaml
config:
  lite:
    enabled: true
    websocket:
      enabled: true
      bind: 0.0.0.0:25567
      path: /mc
      trustedProxies:
        - 127.0.0.1/32
        - 10.0.0.0/8
      forwardedForHeader: X-Forwarded-For
      readLimit: 16777216
      framePayloadSize: 65536
      minFramePayloadSize: 4096
      adaptiveFraming: true
      coalesceWindow: 200us
      coalesceLimit: 65536
      maxPendingBytes: 4194304
      backpressureTimeout: 10s
      idleTimeout: 90s
      tcpNoDelay: true
      tcpKeepAlive: 30s
      tcpKeepAliveInterval: 10s
      tcpKeepAliveCount: 3
      tcpNotSentLowAt: 0
      handshakeTimeout: 10s
      maxConnectionsPerIP: 64
      compression: false
```

客户端连接地址示例：`wss://play.example.com/mc`。

只有连接来源属于 `trustedProxies` 时，Gate 才会信任 `X-Forwarded-For`。解析出的真实 IP 会用于 Gate 限流、日志和 Lite 后端的 PROXY Protocol/TCPShield RealIP。

## 路由模式

### 翻译为 TCP

```yaml
routes:
  - host: play.example.com
    backend: 127.0.0.1:25566
    proxyProtocol: true
    websocket:
      enabled: true
      mode: translate
```

Gate 解包 WebSocket，将连续 Minecraft 字节流交给现有 Lite TCP 转发链。HTTP `Host` 会覆盖 Minecraft 握手虚拟主机，用于域名分流。

### WebSocket raw-passthrough

```yaml
routes:
  - host: edge.example.com
    backend: 127.0.0.1:25568
    websocket:
      enabled: true
      mode: raw-passthrough
      backendScheme: ws
      backendPath: /mc
      backendHost: backend.example.com
```

Gate 建立第二条 WS/WSS 连接，双向转发 Binary message 内容并重新切片。它会向下游发送经过验证的客户端 IP、原 Host 和协议：

- `X-Forwarded-For`
- `X-Forwarded-Host`
- `X-Forwarded-Proto`

`backendHost` 可覆盖下游 HTTP Host；为空时保留玩家请求的 Host。

## 反向代理要求

- CDN/Nginx/Caddy 必须保留 HTTP `Host`。
- 代理地址必须列入 `trustedProxies`，否则 Gate 会忽略其 `X-Forwarded-For`。
- 建议外部使用 WSS，由 CDN 或反向代理终止 TLS；Gate 独立监听端口可以使用内部 WS。
- `readLimit` 限制单个传入 WebSocket message；`framePayloadSize` 控制 Gate 发出的 Binary message 切片大小。

## 传输优化与客户端兼容性

Gate 会在不改变 WSMC 线协议的前提下合并短写、按队列压力和写入耗时在
`minFramePayloadSize` 到 `framePayloadSize` 之间调整 Binary message 大小，并通过
`maxPendingBytes` 和 `backpressureTimeout` 限制每条连接的待发送内存。关闭连接前会排空
已接受的写入，避免丢失最后一段 Minecraft 数据。

TCP 连接默认启用 `TCP_NODELAY` 和 keepalive；读写 socket buffer 可用
`socketReadBuffer`、`socketWriteBuffer` 调整。Linux 可选择设置 `tcpNotSentLowAt`，值为 0
表示保持系统默认。压缩默认关闭，因为 Minecraft 数据本身常已压缩，重复压缩通常增加 CPU
和延迟；确认业务有效时仍可手动开启。

参考客户端只把 Binary message 内容交给 Minecraft TCP pipeline，且没有 Gate 可协商的
应用层 ACK、序号或多路复用协议。因此这里不增加应用层重传（TCP 已负责可靠传输），也不主动
发送依赖客户端 Pong 的保活帧；`idleTimeout` 与 TCP keepalive 用于清理失活连接。
