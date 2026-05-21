package tcpbrutal

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMbpsToBytesPerSecond(t *testing.T) {
	require.Equal(t, uint64(12_500_000), MbpsToBytesPerSecond(100))
	require.Equal(t, uint64(2_500_000), MbpsToBytesPerSecond(20))
}

func TestOptionsNormalize(t *testing.T) {
	options := Options{Enabled: true, RateBytesPerSecond: 1}
	require.Equal(t, DefaultCwndGain, options.Normalize().CwndGain)

	options.CwndGain = 20
	require.Equal(t, uint32(20), options.Normalize().CwndGain)
}

func TestApplyDisabledNoop(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	require.NoError(t, Apply(client, Options{}))
	require.NoError(t, Apply(client, Options{Enabled: true}))
}
