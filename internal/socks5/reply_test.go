package socks5

import (
	"bytes"
	"net"
	"testing"
)

func TestWriteReply(t *testing.T) {
	tests := []struct {
    name     string
    code     ReplyCode
    addr     *net.TCPAddr
    expected []byte
    wantErr  bool
	}{
		{
			name: "IPv4 success",
			code: ReplySuccess,
			addr: &net.TCPAddr{
				IP:   net.ParseIP("1.2.3.4"),
				Port: 443,
			},
			expected: []byte{
				0x05, 0x00, 0x00, 0x01,
				0x01, 0x02, 0x03, 0x04,
				0x01, 0xBB,
			},
		},
		{
			name: "IPv6 success",
			code: ReplySuccess,
			addr: &net.TCPAddr{
				IP:   net.ParseIP("::1"),
				Port: 8080,
			},
			expected: []byte{
				0x05, 0x00, 0x00, 0x04,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x01,
				0x1F, 0x90,
			},
		},
		{
			name: "Connection refused",
			code: ReplyConnRefused,
			addr: nil,
			expected: []byte{
				0x05, 0x05, 0x00, 0x01,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00,
			},
		},
		{
			name: "General failure",
			code: ReplyGeneralFailure,
			addr: nil,
			expected: []byte{
				0x05, 0x01, 0x00, 0x01,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00,
			},
		},
		{
			name: "Command not supported",
			code: ReplyCmdNotSupported,
			addr: nil,
			expected: []byte{
				0x05, 0x07, 0x00, 0x01,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00,
			},
		},
		{
			name: "Invalid IP",
			code: ReplySuccess,
			addr: &net.TCPAddr{
				IP:   net.IP{0x01, 0x02},
				Port: 443,
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Port too large",
			code: ReplySuccess,
			addr: &net.TCPAddr{
				IP:   net.ParseIP("1.2.3.4"),
				Port: 70000,
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Negative port",
			code: ReplySuccess,
			addr: &net.TCPAddr{
				IP:   net.ParseIP("1.2.3.4"),
				Port: -1,
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := writeReply(&buf, tt.code, tt.addr)

			if (err != nil) != tt.wantErr {
				t.Fatalf("writeReply wantErr failed: %v", err)
			}

			if !bytes.Equal(buf.Bytes(), tt.expected) {
				t.Errorf("reply = % x, expected % x", buf.Bytes(), tt.expected)
			}
		})
	}
}