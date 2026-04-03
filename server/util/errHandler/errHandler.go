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

func LogOnError(err error, msg string) bool {
	if err != nil {
		slog.Error(msg, slog.Any("error", err))
		return true
	}
	return false
}
