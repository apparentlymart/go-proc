package proc

import (
	"context"
	"os"
	"os/signal"
)

// Main is a helper for running a process as main purpose of a program. It
// is an opinionated function primarily aimed at long-running processes inside
// similarly-long-running daemon-type programs.
//
// Main starts a process with the given implementation and then blocks until
// it is completed. If an interrupt signal (os.Interrupt) is recieved while
// the process is running then it will be cancelled via its context. Therefore
// process implementations used with this function should exit cleanly but
// promptly when cancelled.
//
// If the process returns any error when it exits, Main returns that error
// verbatim to be handled by the caller.
func Main(impl Impl) error {
	intC := make(chan os.Signal, 1)
	signal.Notify(intC, os.Interrupt)

	exitC := make(chan struct{})
	var err error

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		err = impl(ctx)
		close(exitC)
	}()

	select {
	case <-intC:
	case <-exitC:
	}

	cancel()
	signal.Stop(intC)
	return err
}
