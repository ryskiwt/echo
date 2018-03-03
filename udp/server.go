package main

import (
	"context"
	"io"
	"net"

	"golang.org/x/sync/errgroup"

	"github.com/pkg/errors"
)

func startServer(ctx context.Context) error {

	udpAddr, err := net.ResolveUDPAddr(network, *server)
	if err != nil {
		return errors.WithStack(err)
	}

	conn, err := net.ListenUDP(network, udpAddr)
	if err != nil {
		return errors.WithStack(err)
	}
	logger.Printf("Listening on %s ...", *server)

	return serveConnection(ctx, conn)
}

func serveConnection(ctx context.Context, conn *net.UDPConn) error {
	ctx, cancel := context.WithCancel(ctx)
	eg, ctx := errgroup.WithContext(ctx)

	//
	// clean up
	//

	eg.Go(func() error {
		defer cancel()

		<-ctx.Done()
		return conn.Close()
	})

	//
	// serving
	//

	eg.Go(func() error {
		defer cancel()

		bs := make([]byte, 4*1024) // 4KB

		for {
			n, remoteAddr, err := conn.ReadFromUDP(bs)
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

			logger.Printf("Server: Got '%s' from remote peer %s\n", string(bs[:n]), remoteAddr)

			if _, err = conn.WriteTo(bs[:n], remoteAddr); err != nil {
				if isDone(ctx) {
					return nil
				}
				return errors.WithStack(err)
			}
		}

	})

	return eg.Wait()
}
