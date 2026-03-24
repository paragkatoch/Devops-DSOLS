package errhandler

import (
	"log"
	"log/slog"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("\n%s: %s", msg, err)
	}
}

func LogOnError(err error, msg string) {
	if err != nil {
		slog.Error(msg, slog.String("error", err.Error()))
	}
}
