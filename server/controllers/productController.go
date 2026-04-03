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

var validate = validator.New()

func ProductController(storage storage.Storage, cfg *config.Config) {
	slog.Info("Hello from product controller")

	// connect to queue
	conn, ch := rabbitmq.New(cfg)

	defer conn.Close()
	defer ch.Close()

	q := rabbitmq.Connect(ch, "product")

	// setup router
	router := http.NewServeMux()
	router.HandleFunc("POST /api/product", CreateProduct(ch, q))
	router.HandleFunc("GET /api/product/{id}", GetProduct(storage))
	router.HandleFunc("GET /api/product", GetProducts(storage))
	router.HandleFunc("POST /api/product/quantity", UpdateProductQuantity(ch, q))

	// setup server
	server := &http.Server{
		Addr:    cfg.HTTPServer.Addr,
		Handler: router,
	}

	serverhandler.Serve(server, func() {
		slog.Info("Product Controller started", slog.String("address", cfg.HTTPServer.Addr))
		err := server.ListenAndServe()
		errhandler.FailOnError(err, "Failed to start server")
	})
}

func CreateProduct(ch *amqp091.Channel, q amqp091.Queue) http.HandlerFunc {
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
		err = rabbitmq.SendMessage(ch, q, event)

		if errhandler.LogOnError(err, "Failed to send event") {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

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
		products, err := storage.GetProducts()

		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, products)
	}
}

func UpdateProductQuantity(ch *amqp091.Channel, q amqp091.Queue) http.HandlerFunc {
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
		// send to queue
		err = rabbitmq.SendMessage(ch, q, event)
		if errhandler.LogOnError(err, "Failed to send event") {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, map[string]string{"success": "queued"})
	}
}
