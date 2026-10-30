package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"socks5-proxy/internal/config"
	"socks5-proxy/internal/server"
	"socks5-proxy/internal/socks5"
	"syscall"
	"time"
)

func main()  {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM, syscall.SIGINT,
	)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	configPath := flag.String(
		"config",
		"configs/config.yaml",
		"path to configuration file",
	)

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("config failed", "error", err)
		os.Exit(1)
	}

	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	handler := socks5.NewHandler(logger, dialer)
	srv := server.New(
		net.JoinHostPort(cfg.Server.Host, fmt.Sprint(cfg.Server.Port)), 
		handler, 
		logger,
	)
	
	if err := srv.Serve(ctx); err != nil {
		logger.Error("serve failed", "error", err)
		os.Exit(1)
	}
}