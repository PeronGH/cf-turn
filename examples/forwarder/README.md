# Cloudflare TURN Forwarder

This is a simple TURN forwarder that forwards TCP and UDP over QUIC, proxied by Cloudflare's TURN service.

## Usage

### Server

```bash
# help message
go run ./examples/forwarder/server

# forward SSH
go run ./examples/forwarder/server -a localhost:22

# forward DNS
go run ./examples/forwarder/server -a 1.1.1.1:53
```

### Client

```bash
# help message
go run ./examples/forwarder/client

# forward SSH
go run ./examples/forwarder/client -l 2222 -r $REMOTE_PORT

# connect to forwarded SSH
ssh -p 2222 localhost

# forward DNS
go run ./examples/forwarder/client -l 5353 -r $REMOTE_PORT

# test DNS
dig @localhost -p 5353 github.com
```
