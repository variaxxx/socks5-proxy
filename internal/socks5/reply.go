package socks5

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type ReplyCode byte

const (
	ReplySuccess          ReplyCode = 0x00
	ReplyGeneralFailure   ReplyCode = 0x01
	ReplyNotAllowed       ReplyCode = 0x02
	ReplyNetworkUnreach   ReplyCode = 0x03
	ReplyHostUnreach      ReplyCode = 0x04
	ReplyConnRefused      ReplyCode = 0x05
	ReplyTTLExpired       ReplyCode = 0x06
	ReplyCmdNotSupported  ReplyCode = 0x07
	ReplyAddrNotSupported ReplyCode = 0x08
)

// RFC1928 reply format
// 05  00  00  01   ...
// VER REP RSV ATYP BND
func writeReply(
	w io.Writer,
	code ReplyCode,
	addr *net.TCPAddr,
) error {
	response := make([]byte, 0, 22)

	response = append(response, Version, byte(code), 0x00)

	if addr == nil {
		response = append(response, byte(AddrIPv4))
		response = append(response, make([]byte, net.IPv4len)...)
	} else if ip4 := addr.IP.To4(); ip4 != nil {
		response = append(response, byte(AddrIPv4))
		response = append(response, ip4...)
	} else if ip6 := addr.IP.To16(); ip6 != nil {
		response = append(response, byte(AddrIPv6))
		response = append(response, ip6...)
	} else {
		return fmt.Errorf("invalid address type")
	}

	port := uint16(0)

	if addr != nil {
		if addr.Port < 0 || addr.Port > 65535 {
			return fmt.Errorf("invalid port: %d", addr.Port)
		}

		port = uint16(addr.Port)
	}

	var portBytes [2]byte
	binary.BigEndian.PutUint16(portBytes[:], port)
	response = append(response, portBytes[:]...)

	n, err := w.Write(response)
	if err != nil {
		return fmt.Errorf("write SOCKS5 reply: %w", err)
	}

	if n != len(response) {
		return io.ErrShortWrite
	}

	return nil
}