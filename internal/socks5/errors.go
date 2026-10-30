package socks5

import (
	"context"
	"errors"
	"syscall"
)

func mapDialError(err error) ReplyCode {
	switch {
		case errors.Is(err, syscall.ECONNREFUSED):
        return ReplyConnRefused

    case errors.Is(err, syscall.ENETUNREACH):
        return ReplyNetworkUnreach

    case errors.Is(err, syscall.EHOSTUNREACH):
        return ReplyHostUnreach

    case errors.Is(err, context.DeadlineExceeded):
        return ReplyGeneralFailure

    default:
        return ReplyGeneralFailure
	}
}

var ErrUnsupportedAddressType = errors.New("unsupported address type")