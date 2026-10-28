//go:build !windows

package main

import (
	"os"
	"os/signal"
	"syscall"
	obj "github.com/theawakener0/Zod/object"
)

func installSignalCleanup() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		sig := <-ch
		obj.TermSignalCleanup()
		if s, ok := sig.(syscall.Signal); ok {
			os.Exit(128 + int(s))
		}
		os.Exit(1)
	}()
}
