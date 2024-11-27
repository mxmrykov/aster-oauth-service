package utils

import (
	"os"
	"os/signal"
	"syscall"
)

func GracefulShutDown() chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	return c
}
