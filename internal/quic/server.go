package quic

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"log"
	"math/big"
	"net"

	"github.com/pkg/errors"
	"github.com/quic-go/quic-go"
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

	return quic.Listen(conn, tlsConf, nil)
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

func ForwardSessionsAsServer(ctx context.Context, ln *quic.Listener, addr string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			session, err := ln.Accept(ctx)
			if err != nil {
				log.Printf("listener error: %v", err)
				continue
			}
			log.Printf("accepted session: %v", session.RemoteAddr())

			go serverSessionHandler(ctx, session, addr)
		}
	}
}

func serverSessionHandler(ctx context.Context, session quic.Connection, addr string) {
	defer func() {
		if err := session.CloseWithError(0, "close"); err != nil {
			log.Printf("session close error: %v", err)
		}
	}()
	for {
		stream, err := session.AcceptStream(ctx)
		if err != nil {
			log.Printf("session error: %v", err)
			break
		}
		log.Printf("accepted stream: %v, from: %v", stream.StreamID(), session.RemoteAddr())
		go serverStreamHandler(ctx, stream, addr)
	}
}

func serverStreamHandler(ctx context.Context, stream quic.Stream, addr string) {
	defer stream.Close()

	rConn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("dial error: %v", err)
		return
	}
	defer rConn.Close()

	exchangeData(ctx, stream, rConn)
}
