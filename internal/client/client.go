package client

import (
	"cf-stun/internal/credentials"
	"net"
	"time"

	"github.com/pion/logging"
	"github.com/pion/turn/v3"
	"github.com/pkg/errors"
)

var serverAddr = "turn.speed.cloudflare.com:50000"

func NewClientConn(realm string) (*turn.Client, net.PacketConn, net.PacketConn, error) {
	creds, err := credentials.Get()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Failed to get credentials")
	}

	conn, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Failed to listen")
	}

	cfg := &turn.ClientConfig{
		STUNServerAddr: serverAddr,
		TURNServerAddr: serverAddr,
		Conn:           conn,
		Username:       creds.Username,
		Password:       creds.Password,
		Realm:          realm,
		LoggerFactory:  logging.NewDefaultLoggerFactory(),
		RTO:            time.Second,
	}

	client, err := turn.NewClient(cfg)
	if err != nil {
		CloseClientConn(client, conn, nil)
		return nil, nil, nil, errors.Wrap(err, "Failed to create client")
	}

	// Start listening on the conn provided.
	err = client.Listen()
	if err != nil {
		CloseClientConn(client, conn, nil)
		return nil, nil, nil, errors.Wrap(err, "Failed to listen")
	}

	// Allocate a relay socket on the TURN server. On success, it
	// will return a net.PacketConn which represents the remote
	// socket.
	relayConn, err := client.Allocate()
	if err != nil {
		CloseClientConn(client, conn, relayConn)
		return nil, nil, nil, errors.Wrap(err, "Failed to allocate")
	}

	err = client.CreatePermission(relayConn.LocalAddr())
	if err != nil {
		CloseClientConn(client, conn, relayConn)
		return nil, nil, nil, errors.Wrap(err, "Failed to create permission")
	}

	return client, conn, relayConn, nil
}

func CloseClientConn(client *turn.Client, conn net.PacketConn, relayConn net.PacketConn) {
	if relayConn != nil {
		relayConn.Close()
	}
	if client != nil {
		client.Close()
	}
	if conn != nil {
		conn.Close()
	}
}
