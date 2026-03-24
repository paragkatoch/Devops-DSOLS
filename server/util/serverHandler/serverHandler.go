package serverhandler

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
)

func Serve(server *http.Server) {
	// make a channel for os signals
	done := make(chan os.Signal, 1)
	// detect terminating signal
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// run server on seperate thread
	go func() {
		slog.Info("order Controller started", slog.String("address", "localhost:9000"))
		err := server.ListenAndServe()
		errhandler.FailOnError(err, "Failed to start server")
	}()

	<-done

	slog.Info("shutting down the server")
	// create a context with 5sec exp
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	// try to cancel to sever for 5 sec otherwise throw error
	err := server.Shutdown(ctx)
	errhandler.FailOnError(err, "Failed to shutdown server")

	slog.Info("server shut down successfully")
}
