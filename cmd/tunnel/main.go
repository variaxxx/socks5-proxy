package main

import (
	"context"
	"log"
	"os/signal"
	"socks5-proxy/internal/socks5"
	"syscall"
)

func main()  {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM, syscall.SIGINT,
	)
	defer cancel()

	if err := socks5.ListenAndServe(ctx, "127.0.0.1:6767"); err != nil {
		log.Fatal(err)
	}
}