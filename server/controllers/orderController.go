package controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/paragkatoch/Devops-DSOLS/internal/rabbitmq"
	"github.com/paragkatoch/Devops-DSOLS/types"
	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
	"github.com/paragkatoch/Devops-DSOLS/util/response"
	serverhandler "github.com/paragkatoch/Devops-DSOLS/util/serverHandler"
)

func OrderController() {
	slog.Info("Hello from order controller")

	// connect to queue
	conn, ch := rabbitmq.New()

	defer conn.Close()
	defer ch.Close()

	q := rabbitmq.Connect(ch, "order")

	// setup router
	router := http.NewServeMux()
	router.HandleFunc("POST /order", func(w http.ResponseWriter, r *http.Request) {
		var order types.Order

		// parse
		if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		// validate
		if err := validator.New().Struct(&order); err != nil {
			validationErr := err.(validator.ValidationErrors)
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validationErr))
			return
		}

		// convert to json
		jsonBody, err := json.Marshal(order)
		errhandler.LogOnError(err, "Failed to marshal body")

		// send to queue
		rabbitmq.SendMessage(ch, q, jsonBody)
	})

	// setup server
	server := &http.Server{
		Addr:    "localhost:9000",
		Handler: router,
	}

	serverhandler.Serve(server)

}
