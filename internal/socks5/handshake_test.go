package socks5

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

func TestHandshakeNoAuth(t *testing.T) {
	unifiedTestCase(
		t,
		[]byte{0x05, 0x01, 0x00},
		[]byte{0x05, 0x00},
		false,
	)
}

func TestHandshakeUnsupportedMethod(t *testing.T) {
	unifiedTestCase(
		t,
		[]byte{0x05, 0x02, 0x01, 0x02},
		[]byte{0x05, 0xFF},
		true,
	)
}

func TestHandshakeNoMethods(t *testing.T) {
	unifiedTestCase(
		t,
		[]byte{0x05, 0x00},
		nil,
		true,
	)
}

func TestHandshakeIncompleteMethods(t *testing.T) {
	unifiedTestCase(
		t,
		[]byte{0x05, 0x02, 0x01},
		nil,
		true,
	)
}

func TestHandshakeIncompleteHeader(t *testing.T) {
	unifiedTestCase(
		t,
		[]byte{0x05},
		nil,
		true,
	)
}

func unifiedTestCase(
	t *testing.T,
	request []byte,
	expected []byte,
	wantError bool,
) {
	t.Helper()

	serverConn, clientConn := net.Pipe()

	defer serverConn.Close()
	defer clientConn.Close()

	if err := clientConn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatal(err)
	}

	handler := NewHandler(nil)
	
	result := make(chan error, 1)

	go func() {
		result <- handler.handshake(serverConn)
	}()

	if _, err := clientConn.Write(request); err != nil {
		t.Fatalf("write handshake: %v", err)
	}

	if expected != nil {
    response := make([]byte, len(expected))

    if _, err := io.ReadFull(clientConn, response); err != nil {
        t.Fatalf("read handshake response: %v", err)
    }

    if !bytes.Equal(response, expected) {
        t.Errorf("expected %x, got %x", expected, response)
    }
	}

	err := <-result
	if (err != nil) != wantError {
		t.Errorf("handshake error = %v, wantError = %t", err, wantError)
	}
}