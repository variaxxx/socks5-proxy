package socks5

import (
	"context"
	"fmt"
	"log"
	"net"
)

type Handler struct {}

func NewHandler() *Handler {
	return &Handler{}
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

	log.Printf("[%s] Client connected", remoteAddr)
	
	if err := h.handshake(conn); err != nil {
		return fmt.Errorf("handshake: %w", err)
	}

	req, err := readRequest(conn)
	if err != nil {
		return fmt.Errorf("read SOCKS5 request: %w", err)
	}

	_ = req

	return nil
}