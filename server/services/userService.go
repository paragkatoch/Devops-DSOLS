package services

import (
	"encoding/json"
	"log/slog"
	"net/http"

	config "github.com/paragkatoch/Devops-DSOLS/internal"
	"github.com/paragkatoch/Devops-DSOLS/internal/rabbitmq"
	"github.com/paragkatoch/Devops-DSOLS/internal/storage"
	"github.com/paragkatoch/Devops-DSOLS/types"
	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func UserService(storage storage.Storage, cfg *config.Config) {
	slog.Info("Hello from user service")

	// connect to queue
	conn, ch := rabbitmq.New(cfg)

	defer conn.Close()
	defer ch.Close()

	q := rabbitmq.Connect(ch, "user")

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(cfg.HTTPServer.Addr, nil)
	}()

	// receive messages from queue
	rabbitmq.ReceiveMessage(ch, q, func(b []byte) {
		// parse message event
		var event types.RabbitEvent
		err := json.Unmarshal(b, &event)
		if errhandler.LogOnError(err, "UserService: Error parsing RabbitEvent") {
			return
		}

		switch event.Type {

		case types.UserCreate:
			// parse message data
			var user types.User
			err := json.Unmarshal(event.Data, &user)
			if errhandler.LogOnError(err, "UserService: Error parsing User") {
				return
			}

			err = storage.CreateUser(user)
			if errhandler.LogOnError(err, "Failed to create User") {
				return
			}
		}
	})

}
