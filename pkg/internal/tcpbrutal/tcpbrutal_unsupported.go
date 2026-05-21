//go:build !linux

package tcpbrutal

import "net"

func apply(net.Conn, Options) error {
	return ErrUnsupported
}
