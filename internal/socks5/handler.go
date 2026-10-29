package socks5

import (
	"context"
	"fmt"
	"log/slog"
	"net"
)

type Handler struct {
	logger *slog.Logger
}

func NewHandler(
	logger *slog.Logger,
) *Handler {
	return &Handler{
		logger: logger,
	}
}

func (h *Handler) Handle(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-done:
			return
		}
	}()

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
    // command not supported at the moment
    return fmt.Errorf("unsupported command: %x", req.Command)
}

	h.logger.Info(
    "SOCKS5 request received",
    "command", req.Command,
    "address", req.Address,
    "port", req.Port,
	)

	_ = req

	return nil
}