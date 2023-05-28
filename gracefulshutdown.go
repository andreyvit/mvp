package mvp

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func InterceptShutdownSignals(shutdown func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGHUP)
	go func() {
		<-c
		signal.Reset()
		log.Println("shutting down, interrupt again to force quit")
		shutdown()
	}()
}

// gracefulShutdown tries to do a graceful shutdown, but abandons the attempt
// and falls back to forceful shutdown after a timeout.
func GracefulShutdown(gracePeriod time.Duration, graceful func(ctx context.Context) error, forceful func()) {
	defer forceful()

	ctx, cancel := context.WithTimeout(context.Background(), gracePeriod)
	defer cancel()

	err := graceful(ctx)
	if err == context.DeadlineExceeded {
		log.Println("WARNING: graceful shutdown timed out")
	} else if err != nil {
		log.Fatalf("** ERROR: graceful shutdown failed: %v", err)
	}
}
