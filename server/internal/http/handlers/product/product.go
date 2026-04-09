package productHandlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/paragkatoch/Devops-DSOLS/internal/storage"
	"github.com/paragkatoch/Devops-DSOLS/types"
	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
	"github.com/paragkatoch/Devops-DSOLS/util/response"
)

var validate = validator.New()

func CreateProduct(publish chan interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var product types.Product

		// parse
		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		// validate
		if err := validate.Struct(&product); err != nil {
			validationErr := err.(validator.ValidationErrors)
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validationErr))
			return
		}

		// convert to json
		jsonBody, err := json.Marshal(product)
		if errhandler.LogOnError(err, "Failed to marshal body") {
			return
		}

		event := types.RabbitEvent{
			Type: types.ProductCreate,
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

		response.WriteJson(w, http.StatusOK, map[string]string{"success": "ok"})
	}
}

func GetProduct(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var product types.Product

		// parse
		id := r.PathValue(("id"))
		slog.Info("getting a product with", slog.String("id", id))

		product, err := storage.GetProduct(id)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, product)
	}
}

func GetProducts(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("[+] Product-Controller: Received get products requests")
		products, err := storage.GetProducts()

		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, products)
		slog.Info("[-] Product-Controller: sent get products response")
	}
}

func UpdateProductQuantity(publish chan interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// request body
		var req struct {
			Id       string `json:"id" validate:"required"`
			Quantity int    `json:"quantity" validate:"required"`
		}

		// parse
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		// validate
		if err := validate.Struct(&req); err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(err.(validator.ValidationErrors)))
			return
		}

		// convert to json
		jsonBody, err := json.Marshal(req)
		if errhandler.LogOnError(err, "Failed to marshal body") {
			return
		}

		event := types.RabbitEvent{
			Type: types.ProductQuantity,
			Data: jsonBody,
		}
		// Avoid blocking request goroutines indefinitely when the publisher is saturated.
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

		response.WriteJson(w, http.StatusOK, map[string]string{"success": "queued"})
	}
}
