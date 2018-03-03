package main

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/pkg/errors"
	"golang.org/x/net/websocket"
)

func startClient(ctx context.Context) error {

	//
	// connect websocket
	//

	config, err := websocket.NewConfig("wss://"+*client+"/echo", "https://"+*client)
	if err != nil {
		return err
	}
	config.TlsConfig = &tls.Config{InsecureSkipVerify: true}
	conn, err := websocket.DialConfig(config)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	eg, ctx := errgroup.WithContext(ctx)

	//
	// for clean up
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
				if _, err := conn.Write(msgBytes); err != nil {
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
