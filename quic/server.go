package main

import (
	"context"
	"crypto/tls"
	"io"
	"net"

	"golang.org/x/sync/errgroup"

	quic "github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/qerr"
	"github.com/pkg/errors"
)

func startServer(ctx context.Context) error {

	cert, err := tls.LoadX509KeyPair("../cert/server.crt", "../cert/server.key")
	if err != nil {
		return errors.WithStack(err)
	}
	tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}}

	lis, err := quic.ListenAddr(*server, &tlsConfig, nil)
	if err != nil {
		return errors.WithStack(err)
	}
	logger.Printf("Listening on %s ...", *server)

	return serveListener(ctx, lis)
}

func serveListener(ctx context.Context, lis quic.Listener) error {
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
			sess, err := lis.Accept()
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
			logger.Printf("Remote peer %s connected", sess.RemoteAddr())

			eg.Go(func() error {
				if err := serveSession(ctx, sess); err != nil {
					return errors.WithStack(err)
				}
				logger.Printf("Remote peer %s disconnected", sess.RemoteAddr())
				return nil
			})
		}
	})

	return eg.Wait()
}

func serveSession(ctx context.Context, sess quic.Session) error {
	ctx, cancel := context.WithCancel(ctx)
	eg, ctx := errgroup.WithContext(ctx)

	//
	// cleanup
	//

	eg.Go(func() error {
		defer cancel()

		select {
		case <-sess.Context().Done():
		case <-ctx.Done():
			if err := sess.Close(nil); err != nil {
				return errors.WithStack(err)
			}
			<-sess.Context().Done()
		}
		return nil
	})

	//
	// serving
	//

	eg.Go(func() error {
		defer cancel()

		for {
			stream, err := sess.AcceptStream()
			if err != nil {
				if isDone(ctx) {
					return nil
				}
				if e, ok := err.(*qerr.QuicError); ok && e.ErrorCode == qerr.PeerGoingAway {
					return nil
				}
				return errors.WithStack(err)
			}
			logger.Printf("Stream %d on remote peer %s opened", stream.StreamID(), sess.RemoteAddr())

			eg.Go(func() error {
				if err := serveStream(ctx, stream); err != nil {
					return errors.WithStack(err)
				}
				logger.Printf("Stream %d on remote peer %s closed", stream.StreamID(), sess.RemoteAddr())
				return nil
			})
		}
	})

	return eg.Wait()
}

func serveStream(ctx context.Context, stream quic.Stream) error {
	ctx, cancel := context.WithCancel(ctx)
	eg, ctx := errgroup.WithContext(ctx)

	//
	// clean up
	//

	eg.Go(func() error {
		defer cancel()

		select {
		case <-stream.Context().Done():
		case <-ctx.Done():
			if err := stream.Close(); err != nil {
				return errors.WithStack(err)
			}
			<-stream.Context().Done()
		}
		return nil
	})

	//
	// serving
	//

	eg.Go(func() error {
		defer cancel()

		bs := make([]byte, 4*1024) // 4KB

		for {
			n, err := stream.Read(bs)
			if err != nil {
				if isDone(ctx) {
					return nil
				}
				if err == io.EOF {
					return nil
				}
				if e, ok := errors.Cause(err).(quic.StreamError); ok && e.Canceled() {
					return nil
				}
				if e, ok := errors.Cause(err).(*qerr.QuicError); ok && e.ErrorCode == qerr.PeerGoingAway {
					return nil
				}
				if e, ok := err.(net.Error); ok && e.Temporary() {
					logger.Printf("skip temporary error: %+v", err)
					continue
				}
				return errors.WithStack(err)
			}

			logger.Printf("Server: Got '%s' on stream %d\n", string(bs[:n]), stream.StreamID())

			if _, err = stream.Write(bs[:n]); err != nil {
				if isDone(ctx) {
					return nil
				}
				if e, ok := errors.Cause(err).(quic.StreamError); ok && e.Canceled() {
					return nil
				}
				if e, ok := errors.Cause(err).(*qerr.QuicError); ok && e.ErrorCode == qerr.PeerGoingAway {
					return nil
				}
				return errors.WithStack(err)
			}
		}

	})

	return eg.Wait()
}
