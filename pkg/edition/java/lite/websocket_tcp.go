package lite

import (
	"errors"
	"net"
	"time"

	"go.minekube.com/gate/pkg/edition/java/lite/config"
)

func tuneWebSocketTCP(conn net.Conn, cfg config.WebSocketListenerConfig) error {
	tcp, ok := conn.(*net.TCPConn)
	if !ok {
		return nil
	}
	var errs []error
	if err := tcp.SetNoDelay(cfg.EffectiveTCPNoDelay()); err != nil {
		errs = append(errs, err)
	}
	if cfg.SocketReadBuffer > 0 {
		if err := tcp.SetReadBuffer(cfg.SocketReadBuffer); err != nil {
			errs = append(errs, err)
		}
	}
	if cfg.SocketWriteBuffer > 0 {
		if err := tcp.SetWriteBuffer(cfg.SocketWriteBuffer); err != nil {
			errs = append(errs, err)
		}
	}
	if err := tcp.SetKeepAliveConfig(net.KeepAliveConfig{
		Enable:   true,
		Idle:     time.Duration(cfg.EffectiveTCPKeepAlive()),
		Interval: time.Duration(cfg.EffectiveTCPKeepAliveInterval()),
		Count:    cfg.EffectiveTCPKeepAliveCount(),
	}); err != nil {
		errs = append(errs, err)
	}
	if cfg.TCPNotSentLowAt > 0 {
		if err := setTCPNotSentLowAt(tcp, cfg.TCPNotSentLowAt); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
