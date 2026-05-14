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

func OrderService(storage storage.Storage, cfg *config.Config) {
	slog.Info("Hello from order service")

	// connect to queue
	conn, ch := rabbitmq.New(cfg)

	defer conn.Close()
	defer ch.Close()

	q := rabbitmq.Connect(ch, "order")

	productQ := rabbitmq.Connect(ch, "product")
	productPublish := rabbitmq.GetPublisher(conn, productQ)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(cfg.HTTPServer.Addr, nil)
	}()

	// receive messages from queue
	rabbitmq.ReceiveMessage(ch, q, func(b []byte) {

		// parse message event
		var event types.RabbitEvent
		err := json.Unmarshal(b, &event)
		if errhandler.LogOnError(err, "OrderService: Error parsing RabbitEvent") {
			return
		}

		switch event.Type {

		case types.OrderCreate:
			// parse message data
			var payload types.OrderCreateEvent

			err := json.Unmarshal(event.Data, &payload)
			// err := json.Unmarshal(event.Data, &order)
			if errhandler.LogOnError(err, "OrderService: Error parsing order") {
				return
			}

			err = storage.CreateOrder(payload.Order, payload.Items)
			if errhandler.LogOnError(err, "Failed to create order") {
				return
			}

			err = pushToProductQueue(productPublish, payload.Order.Id, payload.Items)
			if errhandler.LogOnError(err, "Failed to push to product queue") {
				err = storage.UpdateOrderStatus(payload.Order.Id, types.OrderFailed)
				prometheus.OrdersFailed.Inc()
				errhandler.LogOnError(err, "Failed to update order status")
			}

		case types.OrderStatusUpdate:
			var req struct {
				OrderId string            `json:"order_id" validate:"required"`
				Status  types.OrderStatus `json:"status" validate:"required"`
			}

			err := json.Unmarshal(event.Data, &req)

			if errhandler.LogOnError(err, "ProductService: Error parsing order update") {
				return
			}

			err = storage.UpdateOrderStatus(req.OrderId, req.Status)
			if req.Status == types.OrderCompleted {
				prometheus.OrdersCompleted.Inc()
			} else {
				prometheus.OrdersFailed.Inc()
			}
			if errhandler.LogOnError(err, "Failed to update order status") {
				return
			}
		}
	})
}

func pushToProductQueue(publish chan interface{}, orderId string, orderItems []types.OrderItem) error {
	payload := types.ProductReserveEvent{
		OrderID: orderId,
		Items:   orderItems,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	event := types.RabbitEvent{
		Type: types.ProductQuantityReserve,
		Data: body,
	}

	select {
	case publish <- event:
	default:
		return errors.New("publisher busy")
	}

	return nil
}
