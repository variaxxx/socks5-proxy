package socks5

import (
	"fmt"
	"io"
	"net"
	"time"
)


func (h *Handler) handshake(conn net.Conn) (AuthMethod, error) {
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return AuthNoAccepted, fmt.Errorf("set handshake deadline: %w", err)
	}
	defer conn.SetReadDeadline(time.Time{})

	header := make([]byte, 2)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		return AuthNoAccepted, fmt.Errorf("read handshake header: %w", err)
	}

	if header[0] != Version {
		return AuthNoAccepted, fmt.Errorf("unsupported proto version: %d", header[0])
	}

	if header[1] == 0 {
		return AuthNoAccepted, fmt.Errorf("invalid NMETHODS: %d", header[1])
	}

	methods := make([]byte, header[1])
	_, err = io.ReadFull(conn, methods)
	if err != nil {
		return AuthNoAccepted, fmt.Errorf("read client methods: %w", err)
	}

	var selectedMethod AuthMethod = AuthNoAccepted

	if h.auth != nil {
		if hasMethod(methods, AuthUserPass) {
			selectedMethod = AuthUserPass
		}
	} else {
		if hasMethod(methods, AuthNoRequired) {
			selectedMethod = AuthNoRequired
		}
	}
	
	if err := writeMethodSelection(conn, selectedMethod); err != nil {
		return AuthNoAccepted, err
	}

	if selectedMethod == AuthNoAccepted {
		return  selectedMethod, fmt.Errorf("unsupported auth method")
	}
	return selectedMethod, nil
}


func hasMethod(methods []byte, target AuthMethod) bool {
	for _, b := range methods {
		if AuthMethod(b) == target {
			return true
		}
	}
	return  false
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