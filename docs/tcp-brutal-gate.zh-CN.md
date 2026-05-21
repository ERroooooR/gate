# Gate TCP Brutal 支持备忘

本文记录 Gate 中保留的 TCP Brutal 支持逻辑。

## 关键点

- TCP Brutal 不是 Minecraft 协议扩展，也不是新的代理协议；它是 Linux TCP 拥塞控制内核模块。
- Gate 的支持方式是在 TCP socket 上设置 sockopt：
  - `TCP_CONGESTION = "brutal"`
  - `TCP_BRUTAL_PARAMS = 23301`
- `rate` 使用 bytes/s，配置里的 Mbps 会换算为 `mbps * 1000 * 1000 / 8`。
- `cwnd_gain` 使用十分之一单位，例如 `15` 表示 `1.5x`。

参考项目：https://github.com/apernet/tcp-brutal

## 配置

```yaml
config:
  tcpBrutal:
    enabled: true
    downMbps: 100
    upMbps: 20
    cwndGain: 15
```

- `enabled`: 是否启用 TCP Brutal。
- `downMbps`: Gate 发给玩家连接时使用的目标速率。
- `upMbps`: Gate 发给后端连接时使用的目标速率。
- `cwndGain`: 写入 `TCP_BRUTAL_PARAMS.cwnd_gain`，默认建议 `15`。

## 当前应用位置

- 玩家入口 TCP 连接：`pkg/edition/java/proxy/proxy.go` 的 `(*Proxy).listenAndServe`。
- Lite 后端 TCP 连接：`pkg/edition/java/lite/forward.go` 的 `dialRoute`。
- Full Proxy 后端 TCP 连接：`pkg/edition/java/proxy/server.go` 的 `(*serverConnection).dial`。

如果启用后系统不支持 TCP Brutal，或内核没有加载 `tcp-brutal` 模块，Gate 会记录日志，但不会因为设置失败而直接断开普通连接。

## 代码位置

```text
pkg/internal/tcpbrutal
```

- `tcpbrutal.go`: 平台无关的 `Options`、校验、单位换算和公共入口。
- `tcpbrutal_linux.go`: Linux 上通过 `SyscallConn` 和 `setsockopt` 实际设置 socket。
- `tcpbrutal_unsupported.go`: 非 Linux 平台 no-op 或返回明确的 unsupported error。

公共入口：

```go
type Options struct {
    Enabled            bool
    RateBytesPerSecond uint64
    CwndGain           uint32
}

func Apply(conn net.Conn, options Options) error
```

## 测试建议

- Linux 上加载 `tcp-brutal` 模块后启动 Gate，确认启用配置不会启动失败。
- 使用普通 TCP 玩家连接 Gate，确认登录和游戏流程正常。
- 使用 `ss -ti` 或内核侧观测方式确认连接拥塞控制为 `brutal`。
- 未加载 `tcp-brutal` 模块时启用配置，确认日志能给出清晰错误，且普通连接仍可继续。
- 确认 Windows/macOS 构建仍然通过。
