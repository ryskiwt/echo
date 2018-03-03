package main

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"time"

	"golang.org/x/sync/errgroup"

	quic "github.com/lucas-clemente/quic-go"
	"github.com/pkg/errors"
)

func startClient(ctx context.Context) error {

	//
	// open session & stream
	//

	sess, err := quic.DialAddr(*client, &tls.Config{InsecureSkipVerify: true}, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	ctx, cancel := context.WithCancel(ctx)
	eg, ctx := errgroup.WithContext(ctx)
	defer cancel()

	for i := 0; i < *num; i++ {

		stream, err := sess.OpenStreamSync()
		if err != nil {
			return errors.WithStack(err)
		}

		//
		// for crean up
		//

		eg.Go(func() error {
			defer cancel()

			select {
			case <-stream.Context().Done():
			case <-sess.Context().Done():
				if err := stream.Close(); err != nil {
					return errors.WithStack(err)
				}
				<-stream.Context().Done()

			case <-ctx.Done():
				if err := stream.Close(); err != nil {
					return errors.WithStack(err)
				}
				<-stream.Context().Done()

				if err := sess.Close(nil); err != nil {
					return errors.WithStack(err)
				}
				<-sess.Context().Done()
			}
			return nil
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
					logger.Printf("Client: Sending '%s' on stream %d\n", *msg, stream.StreamID())
					if _, err = stream.Write(msgBytes); err != nil {
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
					n, err := stream.Read(bs)
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
					logger.Printf("Client: Got '%s' on stream %d\n", bs[:n], stream.StreamID())
				}
			}
		})

	}

	//
	// wait
	//

	return eg.Wait()
}
