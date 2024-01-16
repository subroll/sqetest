package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/subroll/sqetest/internal/app"
	"github.com/subroll/sqetest/internal/pkg/log"
	"go.uber.org/zap"
)

func main() {
	idleConsClosed := make(chan struct{})
	server, err := app.NewHTTPServer()
	if err != nil {
		log.Fatal("unable to create the server", zap.Error(err))
	}

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Stop(ctx); err != nil {
			log.Error("server shutdown error", zap.Error(err))
		}
		idleConsClosed <- struct{}{}
		close(idleConsClosed)
	}()

	if err := server.Start(); err != nil {
		log.Error("server start error", zap.Error(err))
	}

	<-idleConsClosed
}
