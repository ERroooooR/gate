//go:build !linux

package lite

import "net"

func setTCPNotSentLowAt(_ *net.TCPConn, _ int) error { return nil }
