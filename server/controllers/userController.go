package controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	config "github.com/paragkatoch/Devops-DSOLS/internal"
	"github.com/paragkatoch/Devops-DSOLS/internal/rabbitmq"
	"github.com/paragkatoch/Devops-DSOLS/internal/storage"
	"github.com/paragkatoch/Devops-DSOLS/types"
	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
	"github.com/paragkatoch/Devops-DSOLS/util/response"
	serverhandler "github.com/paragkatoch/Devops-DSOLS/util/serverHandler"
	"github.com/rabbitmq/amqp091-go"
)

func UserController(storage storage.Storage, cfg *config.Config) {
	slog.Info("Hello from user controller")

	// connect to queue
	conn, ch := rabbitmq.New(cfg)

	defer conn.Close()
	defer ch.Close()

	q := rabbitmq.Connect(ch, "user")

	// setup router
	router := http.NewServeMux()
	router.HandleFunc("POST /api/user", CreateUser(ch, q))
	router.HandleFunc("GET /api/user/{id}/order", GetUserOrders(storage))
	router.HandleFunc("GET /api/user/{id}", GetUser(storage))
	router.HandleFunc("GET /api/user", GetUsers(storage))

	// setup server
	server := &http.Server{
		Addr:    cfg.HTTPServer.Addr,
		Handler: router,
	}

	serverhandler.Serve(server, func() {
		slog.Info("User Controller started", slog.String("address", cfg.HTTPServer.Addr))
		err := server.ListenAndServe()
		errhandler.FailOnError(err, "Failed to start server")
	})
}

func CreateUser(ch *amqp091.Channel, q amqp091.Queue) http.HandlerFunc {
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
		err = rabbitmq.SendMessage(ch, q, event)

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
