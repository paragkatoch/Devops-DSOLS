package controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
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

		order.Status = types.OrderCreated
		order.OrderID = uuid.New().String()
		order.CreatedAt = time.Now()
		order.UpdatedAt = time.Now()

		// convert to json
		jsonBody, err := json.Marshal(order)
		errhandler.LogOnError(err, "Failed to marshal body")

		// send to queue
		rabbitmq.SendMessage(ch, q, jsonBody)

		response.WriteJson(w, http.StatusOK, map[string]string{"success": "ok"})
	})

	// setup server
	server := &http.Server{
		Addr:    "localhost:9000",
		Handler: router,
	}

	serverhandler.Serve(server, func() {
		slog.Info("Order Controller started", slog.String("address", "localhost:9000"))
		err := server.ListenAndServe()
		errhandler.FailOnError(err, "Failed to start server")
	})
}
