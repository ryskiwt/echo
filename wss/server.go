package main

import (
	"context"
	"io"
	"net"
	"net/http"

	"golang.org/x/sync/errgroup"

	"github.com/pkg/errors"
	"golang.org/x/net/websocket"
)

func startServer(ctx context.Context) error {

	m := http.NewServeMux()
	m.Handle("/echo", websocket.Handler(echoHandler))

	s := &http.Server{
		Addr:    *server,
		Handler: m,
	}

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()
		return s.Shutdown(context.Background())
	})

	eg.Go(func() error {
		logger.Printf("Listening on %s ...", *server)
		if err := s.ListenAndServeTLS("../cert/server.crt", "../cert/server.key"); err != nil {
			if isDone(ctx) {
				return nil
			}
			return errors.WithStack(err)
		}
		return nil
	})

	return eg.Wait()
}

func echoHandler(conn *websocket.Conn) {
	defer conn.Close()

	logger.Printf("Remote peer %s connected", conn.RemoteAddr())
	defer logger.Printf("Remote peer %s disconnected", conn.RemoteAddr())

	ctx, cancel := context.WithCancel(conn.Request().Context())
	eg, ctx := errgroup.WithContext(ctx)

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

			if _, err := conn.Write(bs[:n]); err != nil {
				if isDone(ctx) {
					return nil
				}
				return errors.WithStack(err)
			}
		}
	})

	if err := eg.Wait(); err != nil {
		logger.Println(err)
	}

}
