package quic

import (
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

var quicConfig = &quic.Config{
	MaxIdleTimeout:     5 * time.Minute,
	KeepAlivePeriod:    10 * time.Second,
	MaxIncomingStreams: 1 << 32,
	EnableDatagrams:    true,
}

type streamToConnWrapper struct {
	quic.Stream
	localAddr  net.Addr
	remoteAddr net.Addr
}

func (s *streamToConnWrapper) LocalAddr() net.Addr {
	return s.localAddr
}

func (s *streamToConnWrapper) RemoteAddr() net.Addr {
	return s.remoteAddr
}

func wrapStreamAsConn(s quic.Stream, localAddr net.Addr, remoteAddr net.Addr) net.Conn {
	return &streamToConnWrapper{s, localAddr, remoteAddr}
}
