package orderHandlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paragkatoch/Devops-DSOLS/internal/storage"
	"github.com/paragkatoch/Devops-DSOLS/types"
	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
	"github.com/paragkatoch/Devops-DSOLS/util/response"
)

var validate = validator.New()

func CreateOrder(publish chan interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var orderHttp types.OrderHttp

		// parse
		if err := json.NewDecoder(r.Body).Decode(&orderHttp); err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		// validate
		if err := validate.Struct(&orderHttp); err != nil {
			validationErr := err.(validator.ValidationErrors)
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validationErr))
			return
		}

		order := types.Order{
			Id:          uuid.NewString(),
			UserID:      orderHttp.UserID,
			TotalAmount: orderHttp.TotalAmount,
			Currency:    orderHttp.Currency,
			Status:      types.OrderProcessing,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		items := make([]types.OrderItem, 0, len(orderHttp.Items))

		for _, item := range orderHttp.Items {
			items = append(items, types.OrderItem{
				OrderID:   order.Id,
				ProductID: item.ProductID,
				Quantity:  -1 * item.Quantity,
			})
		}

		payload := types.OrderCreateEvent{
			Order: order,
			Items: items,
		}

		// convert to json
		jsonBody, err := json.Marshal(payload)
		if errhandler.LogOnError(err, "Failed to marshal body") {
			return
		}

		event := types.RabbitEvent{
			Type: types.OrderCreate,
			Data: jsonBody,
		}

		// send to queue
		// err = rabbitmq.SendMessage(ch, q, event)
		select {
		case publish <- event:
		default:
			err := response.WriteJson(w, http.StatusTooManyRequests, response.Response{
				Status: response.StatusError,
				Error:  "publisher is busy, try again",
			})
			errhandler.LogOnError(err, "Failed to write overload response")
			return
		}

		// if errhandler.LogOnError(err, "Failed to send event") {
		// 	response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
		// 	return
		// }

		response.WriteJson(w, http.StatusOK, map[string]string{"success": string(order.Status), "orderId": order.Id})
	}
}

func GetOrder(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var order types.Order

		// parse
		id := r.PathValue(("id"))
		slog.Info("getting a order with", slog.String("id", id))

		order, err := storage.GetOrder(id)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, order)
	}
}

func GetOrders(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("[+] Product-Controller: Received get products requests")
		orders, err := storage.GetOrders()

		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, orders)
		slog.Info("[-] Product-Controller: sent get products response")
	}
}
