# socks5-proxy

A lightweight SOCKS5 proxy server written in Go as a learning project.

A modular TCP proxy built with Go's standard networking libraries. It supports optional username/password authentication, TCP CONNECT, bidirectional traffic forwarding, and context-based server shutdown.

## Features

- SOCKS5 TCP CONNECT support
- IPv4, IPv6, and domain name destinations
- Optional username/password authentication
- Bidirectional TCP relay with half-close support
- Context-based shutdown on `SIGINT` or `SIGTERM`
- Configurable connection limits
- Handshake, authentication, request, and upstream connection timeouts
- Structured logging with `slog`
- YAML configuration with unknown-field validation
- Unit and integration tests

**Currently unsupported:** UDP ASSOCIATE and BIND.

## Project Structure

```text
.
├── cmd/
│   └── tunnel/
│       └── main.go
├── internal/
│   ├── config/
│   ├── server/
│   └── socks5/
│       ├── handler.go
│       ├── handshake.go
│       ├── auth.go
│       ├── connect.go
│       ├── errors.go
│       ├── request.go
│       ├── reply.go
│       └── relay.go
├── configs/
│   └── config.yaml
└── go.mod
```

## Getting Started

### Requirements

- Go (version specified in `go.mod`)
- curl (optional, for the usage examples)

### Installation

```bash
git clone https://github.com/variaxxx/socks5-proxy.git
cd socks5-proxy

go mod download
go build -o tunnel ./cmd/tunnel
```

### Configuration

Edit `configs/config.yaml`. For example:

```yaml
server:
  host: "127.0.0.1"
  port: 6767
  max_connections: 100

auth:
  require_auth: true
  username: "admin"
  password: "change-me"
```

- `host` and `port` specify the listening address.
- `max_connections` limits active client connections. Use a positive integer. Excess connections are closed immediately.
- `require_auth` enables username/password authentication. Set it to `false` to allow unauthenticated connections.
- `username` and `password` must each contain 1–255 bytes when authentication is enabled.

The application loads `configs/config.yaml` by default, relative to the working directory. Unknown YAML fields are rejected.

Timeouts are currently defined in code: 5 seconds for each handshake, authentication, and request-reading phase, and 10 seconds for establishing an upstream connection. They cannot be configured through YAML.

The checked-in configuration uses example credentials (`admin` / `secret`). The example above uses `change-me`; use the credentials from your actual configuration when connecting.

### Running

Run from the project root:

```bash
go run ./cmd/tunnel
```

Or run the compiled binary:

```bash
./tunnel
```

To use a different configuration file:

```bash
./tunnel -config /path/to/config.yaml
```

Press `Ctrl+C` to stop the server and close active connections.

## Usage

Connect through the proxy using curl and the example configuration above:

```bash
curl -v \
  --socks5-hostname 127.0.0.1:6767 \
  --proxy-user "admin:change-me" \
  https://example.com
```

For a server configured with `require_auth: false`:

```bash
curl -v \
  --socks5-hostname 127.0.0.1:6767 \
  https://example.com
```

The `--socks5-hostname` option lets the proxy resolve destination hostnames.

## How It Works

1. Accepts an incoming TCP connection and checks the connection limit.
2. Performs the SOCKS5 handshake and selects an authentication method.
3. Validates credentials if authentication is enabled.
4. Parses the client's CONNECT request.
5. Establishes an outbound TCP connection using `net.Dialer`.
6. Returns a SOCKS5 reply.
7. Forwards traffic bidirectionally using `io.Copy`, preserving TCP half-close behavior.
8. Closes connections and releases the connection slot when the session ends or the server shuts down.

Each established CONNECT session uses two TCP connections: one between the client and Tunnel, and another between Tunnel and the destination. The configured limit counts client connections, including those still negotiating a session.

## Testing

Run all tests:

```bash
go test ./...
```

Run with the race detector:

```bash
go test -race ./...
```

Run with verbose output:

```bash
go test -v ./...
```

## Security Considerations

- SOCKS5 username/password authentication does **not** encrypt credentials or traffic.
- Tunnel does not currently provide transport-layer encryption.
- HTTPS application traffic remains protected by HTTPS; proxy authentication itself is still unencrypted.
- Replace example credentials before sharing access. Do not commit real credentials to version control.
- Credentials are currently read from YAML; environment variable and secret manager integration are not implemented.
- Binding the proxy to a public interface requires appropriate authentication, firewall rules, and access controls.
- Connection limits help reduce resource exhaustion but do not replace comprehensive abuse protection.

For local development, bind to `127.0.0.1`.

## Roadmap

- [x] SOCKS5 handshake
- [x] TCP CONNECT
- [x] IPv4, IPv6, and domain support
- [x] Username/password authentication
- [x] Bidirectional TCP relay
- [x] Connection limits
- [x] Context-based shutdown
- [ ] Per-user connection limits
- [ ] Access control rules (ACL)
- [ ] Improved metrics and observability
- [ ] TLS-protected transport
- [ ] UDP ASSOCIATE
- [ ] BIND

## References

- [RFC 1928 — SOCKS Protocol Version 5](https://www.rfc-editor.org/rfc/rfc1928)
- [RFC 1929 — Username/Password Authentication for SOCKS V5](https://www.rfc-editor.org/rfc/rfc1929)
