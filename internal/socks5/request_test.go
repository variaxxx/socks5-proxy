package socks5

import (
	"bytes"
	"reflect"
	"testing"
)

func TestReadRequests(t *testing.T) {
	tests := []struct {
		name string
		input []byte
		want *Request
		wantErr bool
	}{
		{
			name: "IPv4 connect",
			input: []byte{
				0x05, 0x01, 0x00, 0x01,
				0x01, 0x02, 0x03, 0x04,
				0x01, 0xBB,
			},
			want: &Request{
				Command:     CmdConnect,
				Address:     "1.2.3.4",
				Port:        443,
			},
		},
		{
			name: "Domain CONNECT",
			input: []byte{
				0x05, 0x01, 0x00, 0x03,
				0x0B,
				'e', 'x', 'a', 'm', 'p', 'l', 'e',
				'.', 'c', 'o', 'm',
				0x01, 0xBB,
			},
			want: &Request{
				Command:     CmdConnect,
				Address:     "example.com",
				Port:        443,
			},
		},
		{
			name: "Invalid version",
			input: []byte{
				0x04, 0x01, 0x00, 0x01,
			},
			wantErr: true,
		},
		{
			name: "Unsupported address type",
			input: []byte{
				0x05, 0x01, 0x00, 0x09,
			},
			wantErr: true,
		},
		{
			name: "Incomplete IPv4 address",
			input: []byte{
				0x05, 0x01, 0x00, 0x01,
				0x01, 0x02,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.input)

			got, err := readRequest(reader)

			if (err != nil) != tt.wantErr {
				t.Fatalf("unwanted error: %v", err)
			}

			if tt.wantErr {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readRequest = %+v, expected %+v", got, tt.want)
			}
		})
	}
}