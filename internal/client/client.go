package client

import (
	"net"

	"github.com/pion/logging"
	"github.com/pion/turn/v3"
	"github.com/pkg/errors"
)

type ClientConnConfig struct {
	ServerAddr string
	Username   string
	Password   string
	Realm      string
}

func NewClientConn(config *ClientConnConfig) (*turn.Client, net.PacketConn, net.PacketConn, error) {
	conn, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Failed to listen")
	}

	cfg := &turn.ClientConfig{
		STUNServerAddr: config.ServerAddr,
		TURNServerAddr: config.ServerAddr,
		Conn:           conn,
		Username:       config.Username,
		Password:       config.Password,
		Realm:          config.Realm,
		LoggerFactory:  logging.NewDefaultLoggerFactory(),
	}

	client, err := turn.NewClient(cfg)
	if err != nil {
		conn.Close()
		return nil, nil, nil, errors.Wrap(err, "Failed to create client")
	}

	// Start listening on the conn provided.
	err = client.Listen()
	if err != nil {
		conn.Close()
		client.Close()
		return nil, nil, nil, errors.Wrap(err, "Failed to listen")
	}

	// Allocate a relay socket on the TURN server. On success, it
	// will return a net.PacketConn which represents the remote
	// socket.
	relayConn, err := client.Allocate()
	if err != nil {
		conn.Close()
		client.Close()
		return nil, nil, nil, errors.Wrap(err, "Failed to allocate")
	}

	return client, conn, relayConn, nil
}

func CloseClientConn(client *turn.Client, conn net.PacketConn, relayConn net.PacketConn) {
	relayConn.Close()
	client.Close()
	conn.Close()
}
