package socks5

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

type Command byte

const (
	CmdConnect Command = 0x01
	CmdBind Command = 0x02
	CmdUDP Command = 0x03
)

type AddressType byte

const (
	AddrIPv4 AddressType = 0x01
	AddrDomain AddressType = 0x03
	AddrIPv6 AddressType = 0x04
)

type Request struct {
	Command Command
	Address string
	Port uint16
}

// RFC1928 request format:
// 05  01  00  03   ...
// VER CMD RSV ATYP DST (address + port)
func readRequest(conn net.Conn) (*Request, error) {
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return nil, fmt.Errorf("set conn deadline: %w", err)
	}
	defer conn.SetDeadline(time.Time{})

	header := make([]byte, 4)

	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, fmt.Errorf("read req header: %w", err) 
	}

	version := header[0]
	command := header[1]
	reserve := header[2]
	addrType := header[3]

	if version != Version {
		return nil, fmt.Errorf("incompatible version: %d", version)
	}
	if reserve != 0x00 {
		return nil, fmt.Errorf("invalid RSV: %d", reserve)
	}

	var address string

	switch AddressType(addrType) {
	case AddrIPv4:
		buf := make([]byte, net.IPv4len)

		if _, err := io.ReadFull(conn, buf); err != nil {
			return nil, fmt.Errorf("read ipv4 addr: %w", err) 
		}

		address = net.IP(buf).String()
	
	case AddrIPv6:
		buf := make([]byte, net.IPv6len)

		if _, err := io.ReadFull(conn, buf); err != nil {
			return nil, fmt.Errorf("read ipv6 addr: %w", err) 
		}

		address = net.IP(buf).String()

	case AddrDomain:
		var length [1]byte

		if _, err := io.ReadFull(conn, length[:]); err != nil {
			return nil, fmt.Errorf("read domain length: %w", err) 
		}

		if length[0] == 0 {
			return nil, fmt.Errorf("empty domain")
		}

		buf := make([]byte, int(length[0]))

		if _, err := io.ReadFull(conn, buf); err != nil {
			return nil, fmt.Errorf("read domain addr: %w", err) 
		}

		address = string(buf)
	
	default:
		return nil, fmt.Errorf("%w: %d", ErrUnsupportedAddressType, addrType)
	}

	var portBytes [2]byte

	if _, err := io.ReadFull(conn, portBytes[:]); err != nil {
			return nil, fmt.Errorf("read destination port: %w", err)
	}

	port := binary.BigEndian.Uint16(portBytes[:])

	return &Request{
		Command: Command(command),
		Address: address,
		Port: port,
	}, nil
}