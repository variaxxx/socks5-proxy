package socks5

import (
	"context"
	"fmt"
	"net"
	"strconv"
)

func (h *Handler) connect(
	ctx context.Context,
	req *Request,
) (net.Conn, error) {
	target := net.JoinHostPort(
		req.Address,
		strconv.Itoa(int(req.Port)),
	)

	upstream, err := h.dialer.DialContext(ctx, "tcp", target)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", target, err)
	}

	return upstream, nil
}