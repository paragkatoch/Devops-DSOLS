package controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	config "github.com/paragkatoch/Devops-DSOLS/internal"
	"github.com/paragkatoch/Devops-DSOLS/internal/rabbitmq"
	"github.com/paragkatoch/Devops-DSOLS/internal/storage"
	"github.com/paragkatoch/Devops-DSOLS/types"
	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
	"github.com/paragkatoch/Devops-DSOLS/util/response"
	serverhandler "github.com/paragkatoch/Devops-DSOLS/util/serverHandler"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func OrderController(storage storage.Storage, cfg *config.Config) {
	slog.Info("Hello from order controller")

	// connect to queue
	conn, ch := rabbitmq.New(cfg)

	defer conn.Close()
	defer ch.Close()

	q := rabbitmq.Connect(ch, "order")

	// setup router
	router := http.NewServeMux()

	router.Handle("/metrics", promhttp.Handler())
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
		order.Id = uuid.New().String()
		order.CreatedAt = time.Now()
		order.UpdatedAt = time.Now()

		// convert to json
		jsonBody, err := json.Marshal(order)
		if errhandler.LogOnError(err, "Failed to marshal body") {
			return
		}

		// send to queue
		rabbitmq.SendMessage(ch, q, jsonBody)

		response.WriteJson(w, http.StatusOK, map[string]string{"success": "ok"})
	})

	// setup server
	server := &http.Server{
		Addr:    cfg.HTTPServer.Addr,
		Handler: router,
	}

	serverhandler.Serve(server, func() {
		slog.Info("Order Controller started", slog.String("address", cfg.HTTPServer.Addr))
		err := server.ListenAndServe()
		errhandler.FailOnError(err, "Failed to start server")
	})
}
