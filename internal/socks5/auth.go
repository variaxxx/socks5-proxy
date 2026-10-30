package socks5

import (
	"crypto/subtle"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	AuthVersion byte = 0x01
)

type Authenticator interface {
	Authenticate(username, password string) bool
}

type StaticAuthenticator struct {
	username string
	password string
}


func NewStaticAuthenticator(
	username, password string,
) *StaticAuthenticator {
	return &StaticAuthenticator{
		username: username,
		password: password,
	}
}


func (a *StaticAuthenticator) Authenticate(
	username, password string,
) bool {
	userMatch := subtle.ConstantTimeCompare([]byte(username), []byte(a.username))
	passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(a.password))

	return userMatch == 1 && passwordMatch == 1
}


func (h *Handler) authenticate(conn net.Conn) error {
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return  fmt.Errorf("set auth deadline: %w", err)
	}
	defer conn.SetDeadline(time.Time{})

	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return fmt.Errorf("read auth header: %w", err)
	}

	if header[0] != AuthVersion {
		return fmt.Errorf("unsupported auth version: %d", header[0])
	}

	unamelen := int(header[1])
	if unamelen == 0 {
		return fmt.Errorf("empty username")
	}

	uname := make([]byte, unamelen)
	if _, err := io.ReadFull(conn, uname); err != nil {
		return fmt.Errorf("read auth username: %w", err)
	}

	var plenBuf [1]byte
	if _, err := io.ReadFull(conn, plenBuf[:]); err != nil {
		return fmt.Errorf("read password len: %w", err)
	}

	plen := int(plenBuf[0])
	if plen == 0 {
		return fmt.Errorf("empty password")
	}

	passwd := make([]byte, plen)
	if _, err := io.ReadFull(conn, passwd); err != nil {
		return fmt.Errorf("read auth password: %w", err)
	}

	if h.auth == nil {
		return fmt.Errorf("authenticator not configured")
	}

	valid := h.auth.Authenticate(string(uname), string(passwd))

	status := byte(0x01)
	if valid {
		status = 0x00
	} 

	n, err := conn.Write([]byte{0x01, status})
	
	if err != nil {
		return fmt.Errorf("write auth response: %w", err)
	}

	if n != 2 {
		return fmt.Errorf("write auth response: %w", io.ErrShortWrite)
	}

	if !valid {
		return fmt.Errorf("invalid credentials")
	}

	return nil
}