package quic

import (
	"context"
	"encoding/binary"

	"github.com/pkg/errors"
)

var ErrInvalidDatagram = errors.New("invalid multiplex datagram")

// byte 0-3: channel id
// byte 4-end: data
type multiplexDatagram []byte

func (p multiplexDatagram) IsInvalid() bool {
	return len(p) < 4
}

func (p multiplexDatagram) ChannelID() uint32 {
	return binary.BigEndian.Uint32(p[:4])
}

func (p multiplexDatagram) Data() []byte {
	return p[4:]
}

func newMultiplexDatagram(channelID uint32, data []byte) multiplexDatagram {
	p := make(multiplexDatagram, 4+len(data))
	binary.BigEndian.PutUint32(p[:4], channelID)
	copy(p[4:], data)
	return p
}

type datagramConn interface {
	SendDatagram(payload []byte) error
	ReceiveDatagram(context.Context) ([]byte, error)
}

type multiplexDatagramConn struct {
	conn datagramConn
}

func newMultiplexDatagramConn(conn datagramConn) *multiplexDatagramConn {
	return &multiplexDatagramConn{conn: conn}
}

func (c *multiplexDatagramConn) Send(channelID uint32, payload []byte) error {
	return c.conn.SendDatagram(newMultiplexDatagram(channelID, payload))
}

func (c *multiplexDatagramConn) Receive(ctx context.Context) (uint32, []byte, error) {
	datagram, err := c.conn.ReceiveDatagram(ctx)
	if err != nil {
		return 0, nil, err
	}
	p := multiplexDatagram(datagram)
	if p.IsInvalid() {
		return 0, nil, ErrInvalidDatagram
	}
	return p.ChannelID(), p.Data(), nil
}
