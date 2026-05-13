package controllers

import (
	"log/slog"
	"net/http"

	config "github.com/paragkatoch/Devops-DSOLS/internal"
	handler "github.com/paragkatoch/Devops-DSOLS/internal/http/handlers/order"
	pt "github.com/paragkatoch/Devops-DSOLS/internal/prometheus"
	"github.com/paragkatoch/Devops-DSOLS/internal/rabbitmq"
	"github.com/paragkatoch/Devops-DSOLS/internal/storage"
	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
	serverhandler "github.com/paragkatoch/Devops-DSOLS/util/serverHandler"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func OrderController(storage storage.Storage, cfg *config.Config) {
	slog.Info("Hello from order controller")

	// connect to queue
	conn, ch := rabbitmq.New(cfg)
	q := rabbitmq.Connect(ch, "order")
	publish := rabbitmq.GetPublisher(conn, q)

	defer conn.Close()
	defer ch.Close()

	// setup router
	router := http.NewServeMux()

	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("POST /api/order", pt.Instrument(handler.CreateOrder(publish), "order", "POST /api/order"))
	router.HandleFunc("GET /api/order/{id}", pt.Instrument(handler.GetOrder(storage), "order", "GET /api/order/{id}"))
	router.HandleFunc("GET /api/order", pt.Instrument(handler.GetOrders(storage), "order", "GET /api/order"))

	// setup server
	server := &http.Server{
		Addr:    cfg.HTTPServer.Addr,
		Handler: router,
		// ReadHeaderTimeout: 5 * time.Second,
		// ReadTimeout:       10 * time.Second,
		// WriteTimeout:      10 * time.Second,
		// IdleTimeout:       60 * time.Second,
	}

	serverhandler.Serve(server, func() {
		slog.Info("Order Controller started", slog.String("address", cfg.HTTPServer.Addr))
		err := server.ListenAndServe()
		errhandler.FailOnError(err, "Failed to start server")
	})
}
