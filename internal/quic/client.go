package quic

import (
	"context"
	"crypto/tls"
	"log"
	"net"

	"github.com/quic-go/quic-go"
)

func NewClientSession(ctx context.Context, conn net.PacketConn, remotePort int) (quic.Connection, error) {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"forwarder"},
	}

	remoteAddr := &net.UDPAddr{
		IP:   conn.LocalAddr().(*net.UDPAddr).IP,
		Port: remotePort,
	}

	return quic.Dial(ctx, conn, remoteAddr, tlsConf, nil)
}

func ForwardSessionAsClient(ctx context.Context, session quic.Connection, localPort int) {
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{Port: localPort})
	if err != nil {
		log.Printf("failed to listen: %v", err)
	}
	defer ln.Close()
	log.Printf("listening on: %s", ln.Addr().String())

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("failed to accept: %v", err)
		}
		log.Printf("accepted connection: %v", conn.RemoteAddr())
		go handleClientConn(ctx, conn, session)
	}
}

func handleClientConn(ctx context.Context, conn net.Conn, session quic.Connection) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("connection close error: %v", err)
		}
	}()

	stream, err := session.OpenStreamSync(ctx)
	if err != nil {
		log.Printf("failed to open stream: %v", err)
	}
	defer func() {
		if err := stream.Close(); err != nil {
			log.Printf("stream close error: %v", err)
		}
	}()

	exchangeData(ctx, conn, stream)
}
