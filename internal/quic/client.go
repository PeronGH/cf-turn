package quic

import (
	"context"
	"crypto/tls"
	"log"
	"net"

	"github.com/quic-go/quic-go"
	"github.com/sagernet/sing/common/bufio"
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

	return quic.Dial(ctx, conn, remoteAddr, tlsConf, quicConfig)
}

func ForwardSessionAsClient(ctx context.Context, session quic.Connection, localPort int) {
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{Port: localPort})
	if err != nil {
		log.Printf("failed to listen: %v", err)
	}
	defer ln.Close()
	log.Printf("listening on %v", ln.Addr())

	connCh := make(chan net.Conn)
	errCh := make(chan error)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				conn, err := ln.Accept()
				if err != nil {
					errCh <- err
					return
				}
				connCh <- conn
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case err := <-errCh:
			log.Printf("failed to accept: %v", err)
		case conn := <-connCh:
			log.Printf("accepted connection: %v", conn.RemoteAddr())
			go handleClientConn(ctx, conn, session)
		}
	}
}

func handleClientConn(ctx context.Context, conn net.Conn, session quic.Connection) {
	stream, err := session.OpenStreamSync(ctx)
	if err != nil {
		log.Printf("failed to open stream: %v", err)
		return
	}

	err = bufio.CopyConn(ctx, conn, wrapStreamAsConn(stream, nil, nil))
	if err != nil {
		log.Printf("copy error: %v", err)
	}
	log.Printf("exchange data finished for %v", conn.RemoteAddr())
}
