package socks5

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
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
	header := make([]byte, 4)

	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, fmt.Errorf("read req header: %w", err) 
	}

	VER := header[0]
	CMD := header[1]
	RSV := header[2]
	ATYP := header[3]

	if VER != Version {
		return nil, fmt.Errorf("incompatible version: %d", VER)
	}
	if RSV != 0x00 {
		return nil, fmt.Errorf("invalid RSV: %d", RSV)
	}

	// only tcp connect command is supported at the moment
	if Command(CMD) != CmdConnect {
		return nil, fmt.Errorf("invalid command: %x", CMD)
	}

	var address string
	var port uint16

	switch AddressType(ATYP) {
	case AddrIPv4:
		buf := make([]byte, net.IPv4len)

		if _, err := io.ReadFull(conn, buf); err != nil {
			return nil, fmt.Errorf("read ipv4 addr: %w", err) 
		}

		address = net.IP(buf).String()
		port = binary.BigEndian.Uint16(buf[len(buf) - 2:])
	
	case AddrIPv6:
		buf := make([]byte, net.IPv6len)

		if _, err := io.ReadFull(conn, buf); err != nil {
			return nil, fmt.Errorf("read ipv6 addr: %w", err) 
		}

		address = net.IP(buf).String()
		port = binary.BigEndian.Uint16(buf[len(buf) - 2:])

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
		port = binary.BigEndian.Uint16(buf[len(buf) - 2:])
	
	default:
		return nil, fmt.Errorf("unsupported address type: %d", ATYP)
	}

	return &Request{
		Command: CmdConnect,
		Address: address,
		Port: port,
	}, nil
}