package quic

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/quic-go/quic-go"
)

var quicConfig = &quic.Config{
	KeepAlivePeriod:    10 * time.Second,
	MaxIncomingStreams: 1 << 32,
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (cr *contextReader) Read(p []byte) (n int, err error) {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		n, err = cr.r.Read(p)
	}()

	select {
	case <-cr.ctx.Done():
		return 0, cr.ctx.Err()
	case <-ch:
		return n, err
	}
}

type contextWriter struct {
	ctx context.Context
	w   io.Writer
}

func (cw *contextWriter) Write(p []byte) (n int, err error) {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		n, err = cw.w.Write(p)
	}()

	select {
	case <-cw.ctx.Done():
		return 0, cw.ctx.Err()
	case <-ch:
		return n, err
	}
}

func exchangeData(ctx context.Context, rw1, rw2 io.ReadWriter) {
	ctx, cancel := context.WithCancel(ctx)

	errCh := make(chan error)
	defer close(errCh)

	go readAndWrite(ctx, rw1, rw2, errCh)
	go readAndWrite(ctx, rw2, rw1, errCh)

	if err := <-errCh; err != nil {
		log.Printf("readAndWrite error: %v", err)
	}
	cancel()
	<-errCh
}

func readAndWrite(ctx context.Context, r io.Reader, w io.Writer, errCh chan<- error) {
	ctxR := &contextReader{ctx: ctx, r: r}
	ctxW := &contextWriter{ctx: ctx, w: w}
	_, err := io.Copy(ctxW, ctxR)
	errCh <- err
}
