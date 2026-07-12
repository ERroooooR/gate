//go:build linux

package lite

import (
	"net"

	"golang.org/x/sys/unix"
)

func setTCPNotSentLowAt(conn *net.TCPConn, value int) error {
	raw, err := conn.SyscallConn()
	if err != nil {
		return err
	}
	var sockErr error
	if err = raw.Control(func(fd uintptr) {
		sockErr = unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_NOTSENT_LOWAT, value)
	}); err != nil {
		return err
	}
	if sockErr != nil {
		return sockErr
	}
	return nil
}
