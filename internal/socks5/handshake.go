package socks5

import (
	"fmt"
	"io"
	"net"
	"time"
)


func (h *Handler) handshake(conn net.Conn) error {
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return fmt.Errorf("set handshake deadline: %w", err)
	}
	defer conn.SetReadDeadline(time.Time{})

	header := make([]byte, 2)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		return fmt.Errorf("read handshake header: %w", err)
	}

	if header[0] != Version {
		return fmt.Errorf("unsupported proto version: %d", header[0])
	}

	if header[1] == 0 {
		return fmt.Errorf("invalid NMETHODS: %d", header[1])
	}

	methods := make([]byte, header[1])
	_, err = io.ReadFull(conn, methods)
	if err != nil {
		return fmt.Errorf("read client methods: %w", err)
	}

	for _, b := range methods {
		if AuthMethod(b) == AuthNoRequired {
			return writeMethodSelection(conn, AuthNoRequired)
		}
	}
	if err := writeMethodSelection(conn, AuthNoAccepted); err != nil {
		return err
	}

	return fmt.Errorf("no supported methods")
}


func writeMethodSelection(conn net.Conn, method AuthMethod) error {
	response := []byte{Version, byte(method)}

	n, err := conn.Write(response)
	if err != nil {
		return fmt.Errorf("write method selection: %w", err)
	}

	if n != len(response) {
		return io.ErrShortWrite
	}

	return nil
}