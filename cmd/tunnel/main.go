package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main()  {
	ctx, cancel := signal.NotifyContext(context.Background(),
	syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	listener, err := net.Listen("tcp", "127.0.0.1:6767")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating listener: %v", err)
		os.Exit(1)
	}
	defer listener.Close()
	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	log.Println("Started accepting connections...")

	var wg sync.WaitGroup

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				log.Println("Server stopped")
				break
			}

			log.Printf("Accept error: %v", err)
			continue
		}
		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := handleConnection(ctx, conn); err != nil {
				log.Printf("[%s] %v", conn.RemoteAddr(), err)
			}
		}()
	}
	
	wg.Wait()
}

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