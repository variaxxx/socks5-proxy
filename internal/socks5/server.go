package socks5

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

func ListenAndServe(ctx context.Context, addr string) error {
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
	return nil
}