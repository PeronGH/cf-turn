# Cloudflare TURN Forwarder

This is a simple TURN forwarder that forwards TCP connections over QUIC, proxied using Cloudflare's TURN service.

## Usage

### Server

```bash
# help message
go run ./examples/forwarder/server

# forward SSH
go run ./examples/forwarder/server -a localhost:22
```

### Client

```bash
# help message
go run ./examples/forwarder/client

# forward SSH
go run ./examples/forwarder/client -r $REMOTE_PORT -l 2222

# connect to forwarded SSH
ssh -p 2222 localhost
```