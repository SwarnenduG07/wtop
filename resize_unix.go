//go:build unix

package main

import (
	"os"
	"os/signal"
	"syscall"
)

func setupResizeCh() <-chan struct{} {
	ch := make(chan struct{}, 1)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)
	go func() {
		for range sigCh {
			select {
			case ch <- struct{}{}:
			default:
			}
		}
	}()
	return ch
}
