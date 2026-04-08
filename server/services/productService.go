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

func ProductService(storage storage.Storage, cfg *config.Config) {
	slog.Info("Hello from product service")

	// connect to queue
	conn, ch := rabbitmq.New(cfg)

	defer conn.Close()
	defer ch.Close()

	q := rabbitmq.Connect(ch, "product")

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(cfg.HTTPServer.Addr, nil)
	}()

	// receive messages from queue
	rabbitmq.ReceiveMessage(ch, q, func(b []byte) {
		// parse message event
		var event types.RabbitEvent
		err := json.Unmarshal(b, &event)
		if errhandler.LogOnError(err, "ProductService: Error parsing RabbitEvent") {
			return
		}

		switch event.Type {

		case types.ProductCreate:
			// parse message data
			var product types.Product
			err := json.Unmarshal(event.Data, &product)
			if errhandler.LogOnError(err, "ProductService: Error parsing product") {
				return
			}

			err = storage.CreateProduct(product)
			if errhandler.LogOnError(err, "Failed to create product") {
				return
			}

		case types.ProductQuantity:
			var req struct {
				Id       string `json:"id" validate:"required"`
				Quantity int    `json:"quantity" validate:"required"`
			}

			err := json.Unmarshal(event.Data, &req)

			if errhandler.LogOnError(err, "ProductService: Error parsing product") {
				return
			}

			err = storage.UpdateProductQuantity(req.Id, req.Quantity)
			if errhandler.LogOnError(err, "Failed to create product") {
				return
			}
		}
	})
}
