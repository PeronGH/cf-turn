package quic

import (
	"context"
	"errors"
	"io"
	"log"
	"sync"
)

func exchangeData(ctx context.Context, rw1, rw2 io.ReadWriter) {
	ctx, cancel := context.WithCancel(ctx)

	var wg sync.WaitGroup
	wg.Add(2)
	c1 := readAndWrite(ctx, rw1, rw2, &wg)
	c2 := readAndWrite(ctx, rw2, rw1, &wg)
	select {
	case err := <-c1:
		if err != nil {
			log.Printf("readAndWrite error on c1: %v", err)
			cancel()
			return
		}
	case err := <-c2:
		if err != nil {
			log.Printf("readAndWrite error on c2: %v", err)
			cancel()
			return
		}
	}
	cancel()
	wg.Wait()
}

// The following code is adapted from https://github.com/moul/quicssh/blob/master/main.go

func readAndWrite(ctx context.Context, r io.Reader, w io.Writer, wg *sync.WaitGroup) <-chan error {
	c := make(chan error)
	go func() {
		if wg != nil {
			defer wg.Done()
		}
		buff := make([]byte, 32*1024)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				nr, err := r.Read(buff)
				if err != nil {
					if !errors.Is(err, io.EOF) {
						c <- err
					}
					return
				}
				if nr > 0 {
					_, err := w.Write(buff[:nr])
					if err != nil {
						if !errors.Is(err, io.EOF) {
							c <- err
						}
						return
					}
				}
			}
		}
	}()
	return c
}
