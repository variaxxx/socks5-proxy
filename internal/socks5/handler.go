package socks5

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
)

func handleConnection(ctx context.Context, conn net.Conn) error {
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
	buffer := make([]byte, 1024)
	totalReceived := 0

	defer func() {
		log.Printf("[%s] Disconnected. Total: %d bytes", remoteAddr, totalReceived)
	}()

	for {
		n, err := conn.Read(buffer)

		if n > 0 {
			totalReceived += n
			log.Printf("[%s] Received: %d bytes", remoteAddr, n)

			writtenN, err := conn.Write(buffer[:n])
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}

				return fmt.Errorf("Write error (sent %d/%d bytes): %w", writtenN, n, err)
			}
		}

		if err != nil {
			if err == io.EOF || ctx.Err() != nil{
				return nil
			}

			return fmt.Errorf("Read error: %w", err)
		}
	}
}