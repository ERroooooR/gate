package lite

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/go-logr/logr"
	"go.minekube.com/gate/pkg/edition/java/lite/config"
	"go.minekube.com/gate/pkg/edition/java/netmc"
	"go.minekube.com/gate/pkg/edition/java/proto/packet"
	"go.minekube.com/gate/pkg/edition/java/proto/state"
	"go.minekube.com/gate/pkg/edition/java/proto/state/states"
	"go.minekube.com/gate/pkg/gate/proto"
	"go.minekube.com/gate/pkg/internal/raknetify/raknet"
	"go.minekube.com/gate/pkg/internal/tcpbrutal"
	"go.minekube.com/gate/pkg/util/errs"
)

type RaknetifyOptions struct {
	Bind             string
	Routes           func() []config.Route
	DialTimeout      time.Duration
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	CompressionLevel int
	StrategyManager  *StrategyManager
	BackendTCPBrutal func() tcpbrutal.Options
	Logger           logr.Logger
}

// HasRaknetifyRoutes reports whether any Lite route has Raknetify enabled.
func HasRaknetifyRoutes(routes []config.Route) bool {
	for _, route := range routes {
		if route.Raknetify.Enabled {
			return true
		}
	}
	return false
}

// HasRawRaknetifyRoutes reports whether any Lite route uses raw Raknetify UDP passthrough.
func HasRawRaknetifyRoutes(routes []config.Route) bool {
	for _, route := range routes {
		if route.Raknetify.Enabled && route.RaknetifyMode() == config.RaknetifyModeRawPassthrough {
			return true
		}
	}
	return false
}

// HasFramedRaknetifyRoutes reports whether any Lite route needs Gate to terminate RakNet frames.
func HasFramedRaknetifyRoutes(routes []config.Route) bool {
	for _, route := range routes {
		if route.Raknetify.Enabled && route.RaknetifyMode() != config.RaknetifyModeRawPassthrough {
			return true
		}
	}
	return false
}

// ServeRaknetify starts a RakNet listener for Gate Lite Raknetify clients.
func ServeRaknetify(ctx context.Context, opts RaknetifyOptions) error {
	if opts.Routes == nil {
		return fmt.Errorf("raknetify routes provider is nil")
	}
	if opts.StrategyManager == nil {
		opts.StrategyManager = NewStrategyManager()
	}
	if opts.BackendTCPBrutal == nil {
		opts.BackendTCPBrutal = func() tcpbrutal.Options { return tcpbrutal.Options{} }
	}

	lc := raknet.ListenConfig{ErrorLog: slog.New(slog.NewTextHandler(io.Discard, nil))}
	ln, err := lc.Listen(opts.Bind)
	if err != nil {
		return err
	}
	defer func() { _ = ln.Close() }()

	log := opts.Logger.WithName("lite").WithName("raknetify").WithValues("bind", ln.Addr())
	log.Info("raknetify lite listener started")

	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		go handleRaknetifyConn(ctx, log, opts, conn)
	}
}

func handleRaknetifyConn(ctx context.Context, log logr.Logger, opts RaknetifyOptions, raw net.Conn) {
	packetConn, ok := raw.(raknetFrameConn)
	if !ok {
		log.Info("closing Raknetify connection with unsupported connection type", "type", fmt.Sprintf("%T", raw))
		_ = raw.Close()
		return
	}
	base := newRaknetifyConn(packetConn)
	connCtx := logr.NewContext(ctx, log.WithValues("clientAddr", raw.RemoteAddr()))
	client, startReadLoop := netmc.NewMinecraftConn(connCtx, base, proto.ServerBound, opts.ReadTimeout, opts.WriteTimeout, opts.CompressionLevel)
	client.SetActiveSessionHandler(state.Handshake, &raknetifyHandshakeHandler{
		conn: client,
		opts: opts,
		log:  log.WithName("handshake"),
	})
	startReadLoop()
}

type raknetifyHandshakeHandler struct {
	conn netmc.MinecraftConn
	opts RaknetifyOptions
	log  logr.Logger
}

func (h *raknetifyHandshakeHandler) HandlePacket(pc *proto.PacketContext) {
	if !pc.KnownPacket() {
		_ = h.conn.Close()
		return
	}
	handshake, ok := pc.Packet.(*packet.Handshake)
	if !ok {
		_ = h.conn.Close()
		return
	}

	nextState := raknetifyStateForProtocol(handshake.NextStatus)
	if nextState == nil {
		_ = h.conn.Close()
		return
	}
	h.conn.SetProtocol(proto.Protocol(handshake.ProtocolVersion))
	h.conn.SetState(nextState)

	switch nextState {
	case state.Login:
		ForwardRaknetify(h.opts.DialTimeout, h.opts.Routes(), h.log, h.conn, handshake, pc, h.opts.StrategyManager, h.opts.BackendTCPBrutal())
	case state.Status:
		h.conn.SetActiveSessionHandler(state.Status, &raknetifyStatusHandler{
			conn:      h.conn,
			opts:      h.opts,
			log:       h.log.WithName("status"),
			handshake: handshake,
			ctx:       pc,
		})
	default:
		_ = h.conn.Close()
	}
}

func (h *raknetifyHandshakeHandler) Disconnected() {}
func (h *raknetifyHandshakeHandler) Activated()    {}
func (h *raknetifyHandshakeHandler) Deactivated()  {}

type raknetifyStatusHandler struct {
	conn      netmc.MinecraftConn
	opts      RaknetifyOptions
	log       logr.Logger
	handshake *packet.Handshake
	ctx       *proto.PacketContext

	receivedRequest bool
}

func (h *raknetifyStatusHandler) HandlePacket(pc *proto.PacketContext) {
	if !pc.KnownPacket() {
		_ = h.conn.Close()
		return
	}
	switch pc.Packet.(type) {
	case *packet.StatusRequest:
		h.handleStatusRequest(pc)
	case *packet.StatusPing:
		defer h.conn.Close()
		if err := h.conn.Write(pc.Payload); err != nil {
			h.log.V(1).Info("error writing StatusPing response", "error", err)
		}
	default:
		_ = h.conn.Close()
	}
}

func (h *raknetifyStatusHandler) handleStatusRequest(pc *proto.PacketContext) {
	if h.receivedRequest {
		_ = h.conn.Close()
		return
	}
	h.receivedRequest = true

	log, res, err := ResolveStatusResponseRaknetify(
		h.opts.DialTimeout,
		h.opts.Routes(),
		h.log,
		h.conn,
		h.handshake,
		h.ctx,
		pc,
		h.opts.StrategyManager,
		h.opts.BackendTCPBrutal(),
	)
	if err != nil {
		errs.V(log, err).Info("could not resolve raknetify lite ping", "error", err)
		_ = h.conn.Close()
		return
	}
	if err := h.conn.WritePacket(res); err != nil {
		h.log.V(1).Info("error writing StatusResponse", "error", err)
	}
}

func (h *raknetifyStatusHandler) Disconnected() {}
func (h *raknetifyStatusHandler) Activated()    {}
func (h *raknetifyStatusHandler) Deactivated()  {}

func raknetifyStateForProtocol(status int) *state.Registry {
	switch states.State(status) {
	case states.StatusState:
		return state.Status
	case states.LoginState:
		return state.Login
	}
	return nil
}
