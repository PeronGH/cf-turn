package quic

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"

	"github.com/PeronGH/datagram-forwarder/forwarder"
	"github.com/charmbracelet/log"
	"github.com/pkg/errors"
	"github.com/quic-go/quic-go"
	"github.com/sagernet/sing/common/bufio"
)

func NewListener(conn net.PacketConn) (*quic.Listener, error) {
	tlsCert, err := generateCert()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate cert")
	}

	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"forwarder"},
	}

	return quic.Listen(conn, tlsConf, quicConfig)
}

// The following code is adapted from https://github.com/moul/quicssh/blob/master/server.go

func generateCert() (tls.Certificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return tls.Certificate{}, err
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tlsCert, nil
}

func ForwardSessionsAsServer(ctx context.Context, ln *quic.Listener, forwarder *forwarder.Server, addr string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		session, err := ln.Accept(ctx)
		if err != nil {
			log.Warnf("listener error: %v", err)
			continue
		}
		log.Infof("accepted session: %v", session.RemoteAddr())

		go serverSessionHandler(ctx, session, addr)
		go func() {
			if err := forwarder.Handle(session); err != nil {
				log.Warnf("error when forwarding UDP from %s: %v", session.RemoteAddr(), err)
			}
		}()
	}
}

func serverSessionHandler(ctx context.Context, session quic.Connection, addr string) {
	defer func() {
		if err := session.CloseWithError(0, "close"); err != nil {
			log.Warnf("session close error: %v", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		stream, err := session.AcceptStream(ctx)
		if err != nil {
			log.Errorf("session error: %v", err)
			break
		}
		log.Infof("accepted stream: %v, from: %v", stream.StreamID(), session.RemoteAddr())
		go serverStreamHandler(ctx, stream, addr)
	}
}

func serverStreamHandler(ctx context.Context, stream quic.Stream, addr string) {
	rConn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Errorf("dial error: %v", err)
		return
	}

	err = bufio.CopyConn(ctx, wrapStreamAsConn(stream, nil, nil), rConn)
	if err != nil {
		log.Warnf("copy error: %v", err)
	}
	log.Infof("exchange data finished for stream: %v", stream.StreamID())
}
