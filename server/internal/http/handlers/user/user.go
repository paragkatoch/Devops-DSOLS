package userHandlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/paragkatoch/Devops-DSOLS/internal/storage"
	"github.com/paragkatoch/Devops-DSOLS/types"
	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
	"github.com/paragkatoch/Devops-DSOLS/util/response"
	"github.com/rabbitmq/amqp091-go"
)

var validate = validator.New()

func CreateUser(ch *amqp091.Channel, publish chan interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user types.User

		// parse
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		// validate
		if err := validate.Struct(&user); err != nil {
			validationErr := err.(validator.ValidationErrors)
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validationErr))
			return
		}

		// convert to json
		jsonBody, err := json.Marshal(user)
		if errhandler.LogOnError(err, "Failed to marshal body") {
			return
		}

		event := types.RabbitEvent{
			Type: types.UserCreate,
			Data: jsonBody,
		}

		// send to queue
		// err = rabbitmq.SendMessage(ch, q, event)
		publish <- event

		if errhandler.LogOnError(err, "Failed to send event") {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, map[string]string{"success": "ok"})
	}
}

func GetUser(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse
		id := r.PathValue(("id"))
		slog.Info("getting user with", slog.String("id", id))

		user, err := storage.GetUser(id)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, user)
	}
}

func GetUsers(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := storage.GetUsers()

		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, users)
	}
}

func GetUserOrders(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse
		id := r.PathValue(("id"))
		slog.Info("getting user orders with", slog.String("userId", id))

		orders, err := storage.GetUserOrders(id)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, orders)
	}

}
