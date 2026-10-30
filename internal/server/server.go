package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

const (
	minDelay = 5 * time.Millisecond
	maxDelay = 1 * time.Second
)

type Server struct {
	addr string
	handler Handler
	logger *slog.Logger
	slots chan struct{}
}

func New(
	addr string,
	handler Handler,
	logger *slog.Logger,
	maxConnections int,
) *Server {
	return &Server{
		addr: addr,
		handler: handler,
		logger: logger,
		slots: make(chan struct{}, maxConnections),
	}
}

func (s *Server) Serve(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("Error creating listener: %w", err)
	}
	defer listener.Close()
	
	serverCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-serverCtx.Done()
		listener.Close()
	}()

	var wg sync.WaitGroup
	var serveErr error
	var delay time.Duration
	
	s.logger.Info("Started accepting connections...", "addr", s.addr)

	for {
		conn, err := listener.Accept()
		if serverCtx.Err() != nil {
			if conn != nil {
				conn.Close()
			}

			s.logger.Info("Server stopped")
			break
		}

		if err != nil {
			if !isRetryableError(err) {
				serveErr = fmt.Errorf("accept connection: %w", err)
				break
			}

			if delay == 0 {
				delay = minDelay
			} else {
				delay = min(delay * 2, maxDelay)
			}

			s.logger.Warn(
				"temporary accept error",
				"error", err,
				"retry_in", delay,
			)

			select {
			case <-time.After(delay):
				continue
			case <-serverCtx.Done():
				break
			}

			break
		}

		select {
		case s.slots <- struct{}{}:
			// there is available slot
		default:
			s.logger.Warn("connections limit reached")

			conn.Close()
			continue
		}

		delay = 0
		wg.Add(1)

		go func() {
			defer wg.Done()
			defer func() {
				<-s.slots
			}()

			if err := s.handler.Handle(serverCtx, conn); err != nil {
				s.logger.Error(
					"connection failed", 
					"remote_addr", conn.RemoteAddr().String(), 
					"error", err,
				)
			}
		}()
	}
	
	cancel()
	wg.Wait()
	return serveErr
}