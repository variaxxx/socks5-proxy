package socks5

type AuthMethod byte

const (
	Version byte = 0x05

	AuthNoRequired AuthMethod = 0x00
	AuthUserPass AuthMethod = 0x02
	AuthNoAccepted AuthMethod = 0xFF 
)