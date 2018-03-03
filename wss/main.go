package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	client   = flag.String("c", "", "client mode : -c <server address>")
	server   = flag.String("s", "", "server mode : -s <server listen address>")
	msg      = flag.String("m", "Hello WebSocket !", "message to send")
	interval = flag.Int("i", 1000, "sending interval [ms]")
)

var logger = log.New(os.Stdout, "[echo-wss]", log.Ldate)

func main() {
	flag.Parse()

	if len(*client) == 0 && len(*server) == 0 {
		logger.Fatalf("client or server address is required")
	}

	if len(*client) != 0 && len(*server) != 0 {
		logger.Fatalf("cant set both of client and server address")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	finChan := make(chan struct{}, 1)
	if len(*server) != 0 {
		go func() {
			defer close(finChan)

			if err := startServer(ctx); err != nil {
				logger.Fatalf("%+v", err)
			}
			cancel()
		}()
	}

	if len(*client) != 0 {
		go func() {
			defer close(finChan)

			if err := startClient(ctx); err != nil {
				logger.Fatalf("%+v", err)
			}
			cancel()
		}()
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	select {
	case <-sigChan:
		cancel()
	case <-ctx.Done():
	}

	<-finChan
}
