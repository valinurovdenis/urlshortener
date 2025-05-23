// Package runner for running service with given config.
package runner

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestRunner_Stop(t *testing.T) {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	stopped := make(chan struct{}, 1)
	go Run(ctx, stopped)
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		panic(err)
	}
	p.Signal(syscall.SIGINT)
	<-stopped
}
