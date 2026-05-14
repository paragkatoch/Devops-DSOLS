package services

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	config "github.com/paragkatoch/Devops-DSOLS/internal"
	"github.com/paragkatoch/Devops-DSOLS/internal/prometheus"
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

	orderQ := rabbitmq.Connect(ch, "order")
	orderPublish := rabbitmq.GetPublisher(conn, orderQ)

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
			if errhandler.LogOnError(err, "Failed to update product quantity") {
				return
			} else {
				prometheus.ProductInventory.
					WithLabelValues(req.Id).
					Set(float64(req.Quantity))
			}

		case types.ProductQuantityReserve:
			var productReserve types.ProductReserveEvent

			err := json.Unmarshal(event.Data, &productReserve)
			if errhandler.LogOnError(err, "ProductService: Error parsing productReserve") {
				return
			}

			err = storage.UpdateProductQuantityTransaction(productReserve.Items)
			if errhandler.LogOnError(err, "Failed to update product quantity") {
				err = pushToOrderQueue(orderPublish, productReserve.OrderID, types.OrderFailed)
				errhandler.LogOnError(err, "Failed to update order status")
			} else {
				err = pushToOrderQueue(orderPublish, productReserve.OrderID, types.OrderCompleted)
				errhandler.LogOnError(err, "Failed to update order status")

				for _, product := range productReserve.Items {
					prometheus.ProductInventory.
						WithLabelValues(product.ProductID).
						Set(float64(product.Quantity))
				}
			}
		}
	})
}

func pushToOrderQueue(publish chan interface{}, orderId string, orderStatus types.OrderStatus) error {

	payload := struct {
		OrderId string            `json:"order_id" validate:"required"`
		Status  types.OrderStatus `json:"status" validate:"required"`
	}{
		OrderId: orderId,
		Status:  orderStatus,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	event := types.RabbitEvent{
		Type: types.OrderStatusUpdate,
		Data: body,
	}

	select {
	case publish <- event:
	default:
		return errors.New("publisher busy")
	}

	return nil
}
