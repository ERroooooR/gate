package tcpbrutal

import (
	"errors"
	"fmt"
	"math"
	"net"
)

const (
	DefaultCwndGain       uint32 = 15
	MinCwndGain           uint32 = 5
	MaxCwndGain           uint32 = 80
	MinRateBytesPerSecond uint64 = 62_500
	CongestionControl            = "brutal"
	ParamsSockopt                = 23301
	bytesPerMegabit              = 1000 * 1000 / 8
)

var (
	ErrUnsupported    = errors.New("tcp brutal is unsupported on this platform")
	ErrNotTCPConn     = errors.New("connection does not expose a TCP syscall connection")
	ErrInvalidOptions = errors.New("invalid TCP Brutal options")
)

type Options struct {
	Enabled            bool
	RateBytesPerSecond uint64
	CwndGain           uint32
}

func MbpsToBytesPerSecond(mbps uint64) uint64 {
	if mbps > math.MaxUint64/bytesPerMegabit {
		return math.MaxUint64
	}
	return mbps * bytesPerMegabit
}

func (o Options) Normalize() Options {
	if o.CwndGain == 0 {
		o.CwndGain = DefaultCwndGain
	}
	return o
}

func (o Options) Validate() error {
	o = o.Normalize()
	if !o.Enabled || o.RateBytesPerSecond == 0 {
		return nil
	}
	if o.RateBytesPerSecond < MinRateBytesPerSecond {
		return fmt.Errorf("%w: rate must be at least %d bytes/s", ErrInvalidOptions, MinRateBytesPerSecond)
	}
	if o.CwndGain < MinCwndGain || o.CwndGain > MaxCwndGain {
		return fmt.Errorf("%w: cwnd gain must be between %d and %d", ErrInvalidOptions, MinCwndGain, MaxCwndGain)
	}
	return nil
}

func Apply(conn net.Conn, options Options) error {
	options = options.Normalize()
	if !options.Enabled || options.RateBytesPerSecond == 0 {
		return nil
	}
	if err := options.Validate(); err != nil {
		return err
	}
	return apply(conn, options)
}
