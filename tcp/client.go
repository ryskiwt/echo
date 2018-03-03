package main

import (
	"context"
	"io"
	"net"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/pkg/errors"
)

func startClient(ctx context.Context) error {

	//
	// open session & stream
	//

	tcpAddr, err := net.ResolveTCPAddr(network, *client)
	if err != nil {
		return errors.WithStack(err)
	}

	conn, err := net.DialTCP(network, nil, tcpAddr)
	if err != nil {
		return errors.WithStack(err)
	}

	ctx, cancel := context.WithCancel(ctx)
	eg, ctx := errgroup.WithContext(ctx)

	//
	// for crean up
	//

	eg.Go(func() error {
		defer cancel()

		<-ctx.Done()
		return conn.Close()
	})

	//
	// sending
	//

	eg.Go(func() error {
		defer cancel()

		msgBytes := []byte(*msg)
		ticker := time.NewTicker(time.Duration(*interval) * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil

			case <-ticker.C:
				logger.Printf("Client: Sending '%s'\n", *msg)
				if _, err = conn.Write(msgBytes); err != nil {
					if isDone(ctx) {
						return nil
					}
					return errors.WithStack(err)
				}
			}
		}
	})

	//
	// receiving
	//

	eg.Go(func() error {
		defer cancel()

		bs := make([]byte, 4*1024)

		for {
			select {
			case <-ctx.Done():
				return nil

			default:
				n, err := conn.Read(bs)
				if err != nil {
					if isDone(ctx) {
						return nil
					}
					if err == io.EOF {
						return nil
					}
					if e, ok := err.(net.Error); ok && e.Temporary() {
						logger.Printf("skip temporary error: %+v", err)
						continue
					}
					return errors.WithStack(err)
				}
				logger.Printf("Client: Got '%s'\n", bs[:n])
			}
		}
	})

	//
	// wait
	//

	return eg.Wait()
}
