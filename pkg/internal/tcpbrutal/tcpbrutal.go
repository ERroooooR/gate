package tcpbrutal

import (
	"errors"
	"net"
)

const (
	DefaultCwndGain   uint32 = 15
	CongestionControl        = "brutal"
	ParamsSockopt            = 23301
	bytesPerMegabit          = 1000 * 1000 / 8
)

var (
	ErrUnsupported = errors.New("tcp brutal is unsupported on this platform")
	ErrNotTCPConn  = errors.New("connection does not expose a TCP syscall connection")
)

type Options struct {
	Enabled            bool
	RateBytesPerSecond uint64
	CwndGain           uint32
}

func MbpsToBytesPerSecond(mbps uint64) uint64 {
	return mbps * bytesPerMegabit
}

func (o Options) Normalize() Options {
	if o.CwndGain == 0 {
		o.CwndGain = DefaultCwndGain
	}
	return o
}

func Apply(conn net.Conn, options Options) error {
	options = options.Normalize()
	if !options.Enabled || options.RateBytesPerSecond == 0 {
		return nil
	}
	return apply(conn, options)
}
