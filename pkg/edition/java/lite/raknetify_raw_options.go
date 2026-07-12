package lite

import (
	"github.com/go-logr/logr"
	"go.minekube.com/gate/pkg/edition/java/lite/config"
)

// RaknetifyRawOptions configures the UDP-only Raknetify passthrough listener.
type RaknetifyRawOptions struct {
	Bind            string
	Routes          func() []config.Route
	StrategyManager *StrategyManager
	Logger          logr.Logger
}

// HasRawRaknetifyRoutes reports whether any Lite route enables raw Raknetify passthrough.
func HasRawRaknetifyRoutes(routes []config.Route) bool {
	for _, route := range routes {
		if route.Raknetify.Enabled && route.RaknetifyMode() == config.RaknetifyModeRawPassthrough {
			return true
		}
	}
	return false
}
