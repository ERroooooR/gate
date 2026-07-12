//go:build linux

package tcpbrutal

import (
	"fmt"
	"net"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

type syscallConn interface {
	SyscallConn() (syscall.RawConn, error)
}

type brutalParams struct {
	Rate     uint64
	CwndGain uint32
}

func apply(conn net.Conn, options Options) error {
	sysConn, ok := conn.(syscallConn)
	if !ok {
		return ErrNotTCPConn
	}

	rawConn, err := sysConn.SyscallConn()
	if err != nil {
		return fmt.Errorf("get syscall connection: %w", err)
	}

	var controlErr error
	if err := rawConn.Control(func(fd uintptr) {
		previous, previousErr := unix.GetsockoptString(int(fd), syscall.IPPROTO_TCP, syscall.TCP_CONGESTION)
		controlErr = syscall.SetsockoptString(int(fd), syscall.IPPROTO_TCP, syscall.TCP_CONGESTION, CongestionControl)
		if controlErr != nil {
			return
		}

		params := brutalParams{
			Rate:     options.RateBytesPerSecond,
			CwndGain: options.CwndGain,
		}
		_, _, errno := syscall.Syscall6(
			syscall.SYS_SETSOCKOPT,
			fd,
			uintptr(syscall.IPPROTO_TCP),
			uintptr(ParamsSockopt),
			uintptr(unsafe.Pointer(&params)),
			unsafe.Sizeof(params),
			0,
		)
		if errno != 0 {
			controlErr = errno
			if previousErr == nil && previous != "" {
				if rollbackErr := syscall.SetsockoptString(int(fd), syscall.IPPROTO_TCP, syscall.TCP_CONGESTION, previous); rollbackErr != nil {
					controlErr = fmt.Errorf("set params: %w; rollback congestion control to %q: %v", errno, previous, rollbackErr)
				}
			}
		}
	}); err != nil {
		return fmt.Errorf("control TCP socket: %w", err)
	}
	if controlErr != nil {
		return fmt.Errorf("set TCP brutal options: %w", controlErr)
	}

	return nil
}
