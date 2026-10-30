package socks5

import (
	"fmt"
	"io"
	"net"
)

type copyResult struct {
	direction string
	bytes int64
	err error
}

type closeWriter interface {
	CloseWrite() error
}

func relay(
	client, upstream net.Conn,
) error {
	resultCh := make(chan copyResult, 2)

	go func() {
		written, err := io.Copy(client, upstream)
		
		if err == nil {
			if tcp, ok := client.(closeWriter); ok {
				if closeErr := tcp.CloseWrite(); closeErr != nil {
					err = fmt.Errorf("close client write: %w", closeErr)
				}
			}
		}
		
		resultCh <- copyResult{
			"to_client",
			written,
			err,
		}
	}()

	go func() {
		written, err := io.Copy(upstream, client)

		if err == nil {
			if tcp, ok := upstream.(closeWriter); ok {
				if closeErr := tcp.CloseWrite(); closeErr != nil {
					err = fmt.Errorf("close upstream write: %w", closeErr)
				}
			}
		}

		resultCh <- copyResult{
			"to_upstream",
			written,
			err,
		}
	}()

	first := <-resultCh
	
	if first.err != nil {
		client.Close()
		upstream.Close()
	}
	
	second := <-resultCh
	
	if first.err != nil {
		return fmt.Errorf("failed %s: %w", first.direction, first.err)
	}
	
	if second.err != nil {
		return fmt.Errorf("failed %s: %w", second.direction, second.err)
	}

	return nil
}