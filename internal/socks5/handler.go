package socks5

import (
	"context"
	"fmt"
	"log/slog"
	"net"
)

type Dialer interface {
	DialContext(
		ctx context.Context,
		network string,
		address string,
	) (net.Conn, error)
}

type Handler struct {
	logger *slog.Logger
	dialer Dialer
}

func NewHandler(
	logger *slog.Logger,
	dialer Dialer,
) *Handler {
	return &Handler{
		logger: logger,
		dialer: dialer,
	}
}

func (h *Handler) Handle(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	stopClient := context.AfterFunc(ctx, func() {
		conn.Close()
	})
	defer stopClient()

	remoteAddr := conn.RemoteAddr().String()

	h.logger.Info("client connected", "remote_addr", remoteAddr)
	
	if err := h.handshake(conn); err != nil {
		return fmt.Errorf("handshake: %w", err)
	}

	req, err := readRequest(conn)
	if err != nil {
		return fmt.Errorf("read SOCKS5 request: %w", err)
	}

	if req.Command != CmdConnect {
    // other commands not supported at the moment
		if err := writeReply(conn, ReplyCmdNotSupported, nil); err != nil {
			return fmt.Errorf("write unsupported command reply: %w", err)
		}

    return fmt.Errorf("unsupported SOCKS5 command: %x", req.Command)
}

	h.logger.Info(
    "SOCKS5 request received",
    "command", req.Command,
    "address", req.Address,
    "port", req.Port,
	)

	upstream, err := h.connect(ctx, req)
	if err != nil {
		if replyErr := writeReply(conn, ReplyGeneralFailure, nil); replyErr != nil {
			return fmt.Errorf("connect: %v; write failure reply: %w", err, replyErr)
		}
		return err
	}
	defer upstream.Close()

	stopUpstream := context.AfterFunc(ctx, func() {
		upstream.Close()
	})
	defer stopUpstream()

	boundAddr, ok := upstream.LocalAddr().(*net.TCPAddr)
	if !ok {
		if err := writeReply(conn, ReplyGeneralFailure, nil); err != nil {
			return fmt.Errorf("write failure reply: %w", err)
		}
		return fmt.Errorf("unexpected local address type: %T", upstream.LocalAddr())
	}

	if err := writeReply(conn, ReplySuccess, boundAddr); err != nil {
		return fmt.Errorf("write success reply: %w", err)
	}

	err = relay(conn, upstream); 
	
	if ctx.Err() != nil {
		return nil
	}

	if err != nil {
		return fmt.Errorf("relay error: %w", err)
	}

	return nil
}