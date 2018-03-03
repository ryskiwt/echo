package main

import (
	"context"
	"io"
	"net"

	"golang.org/x/sync/errgroup"

	"github.com/pkg/errors"
)

func startServer(ctx context.Context) error {

	tcpAddr, err := net.ResolveTCPAddr(network, *server)
	if err != nil {
		return errors.WithStack(err)
	}

	lis, err := net.ListenTCP(network, tcpAddr)
	if err != nil {
		return errors.WithStack(err)
	}
	logger.Printf("Listening on %s ...", *server)

	return serveListener(ctx, lis)
}

func serveListener(ctx context.Context, lis net.Listener) error {
	ctx, cancel := context.WithCancel(ctx)
	eg, ctx := errgroup.WithContext(ctx)

	//
	// clean up
	//

	eg.Go(func() error {
		defer cancel()

		<-ctx.Done()
		return lis.Close()
	})

	//
	// serving
	//

	eg.Go(func() error {
		defer cancel()

		for {
			conn, err := lis.Accept()
			if err != nil {
				if isDone(ctx) {
					return nil
				}
				if e, ok := err.(net.Error); ok && e.Temporary() {
					logger.Printf("skip temporary error: %+v", err)
					continue
				}
				return errors.WithStack(err)
			}
			logger.Printf("Remote peer %s connected", conn.RemoteAddr())

			eg.Go(func() error {
				if err := serveConnection(ctx, conn); err != nil {
					return errors.WithStack(err)
				}
				logger.Printf("Remote peer %s disconnected", conn.RemoteAddr())
				return nil
			})
		}
	})

	return eg.Wait()
}

func serveConnection(ctx context.Context, conn net.Conn) error {
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

			logger.Printf("Server: Got '%s'\n", string(bs[:n]))

			if _, err = conn.Write(bs[:n]); err != nil {
				if isDone(ctx) {
					return nil
				}
				return errors.WithStack(err)
			}
		}

	})

	return eg.Wait()
}
