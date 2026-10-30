package server

import (
	"errors"
	"net"
	"syscall"
)

func isRetryableError(err error) bool {
	var netErr net.Error

	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	return errors.Is(err, syscall.EMFILE) ||
		errors.Is(err, syscall.ENFILE) ||
		errors.Is(err, syscall.EINTR) ||
		errors.Is(err, syscall.ECONNABORTED)
}