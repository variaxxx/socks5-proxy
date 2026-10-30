package socks5

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"
)

func TestHandlerConnect(t *testing.T) {
	target, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()

	go func() {
		conn, err := target.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		io.Copy(conn, conn)
	}()

	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	dialer := &net.Dialer{
		Timeout: time.Second,
	}
	handler := NewHandler(logger, dialer)

	handlerDone := make(chan error, 1)

	go func() {
		handlerDone <- handler.Handle(ctx, server)
	}()

	if err := client.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}

	// handshake
	if _, err := client.Write([]byte{0x05, 0x01, 0x00}); err != nil {
		t.Fatal(err)
	}

	methodReply := make([]byte, 2)
	if _, err := io.ReadFull(client, methodReply); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(methodReply, []byte{0x05, 0x00}) {
		t.Fatalf("unexpected handshake reply: % x", methodReply)
	}

	// tcp connect query
	targetAddr := target.Addr().(*net.TCPAddr)

	request := []byte{
		0x05,
		0x01,
		0x00,
		0x01,
	}

	request = append(request, targetAddr.IP.To4()...)

	var port [2]byte
	binary.BigEndian.PutUint16(port[:], uint16(targetAddr.Port))
	request = append(request, port[:]...)

	if _, err := client.Write(request); err != nil {
		t.Fatal(err)
	}

	// verify connect reply
	reply := make([]byte, 10)

	if _, err := io.ReadFull(client, reply); err != nil {
		t.Fatal(err)
	}

	if reply[0] != Version || reply[1] != byte(ReplySuccess) {
		t.Fatalf("unexpected reply: % x", reply)
	}

	if reply[3] != byte(AddrIPv4) {
		t.Fatalf("unexpected address type: %x", reply[3])
	}

	// test duplex transport
	payload := []byte("hello")

	if _, err := client.Write(payload); err != nil {
		t.Fatal(err)
	}

	received := make([]byte, len(payload))

	if _, err := io.ReadFull(client, received); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(payload, received) {
		t.Errorf("received: %q, expected: %q", received, payload)
	}

	// shutdown
	cancel()
	client.Close()

	select {
	case <-handlerDone:
	case <-time.After(2 * time.Second):
		t.Fatal("handler didnt stop after cancellation")
	}
}