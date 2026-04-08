package controllers

import (
	"log/slog"
	"net/http"

	config "github.com/paragkatoch/Devops-DSOLS/internal"
	handler "github.com/paragkatoch/Devops-DSOLS/internal/http/handlers/product"
	pt "github.com/paragkatoch/Devops-DSOLS/internal/prometheus"
	"github.com/paragkatoch/Devops-DSOLS/internal/rabbitmq"
	"github.com/paragkatoch/Devops-DSOLS/internal/storage"
	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
	serverhandler "github.com/paragkatoch/Devops-DSOLS/util/serverHandler"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func ProductController(storage storage.Storage, cfg *config.Config) {
	slog.Info("Hello from product controller")

	// connect to queue
	conn, ch := rabbitmq.New(cfg)
	q := rabbitmq.Connect(ch, "product")
	defer conn.Close()
	defer ch.Close()

	// setup router
	router := http.NewServeMux()

	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("POST /api/product", pt.Instrument(handler.CreateProduct(ch, q), "product", "POST /api/product"))
	router.HandleFunc("GET /api/product/{id}", pt.Instrument(handler.GetProduct(storage), "product", "GET /api/product/{id}"))
	router.HandleFunc("GET /api/product", pt.Instrument(handler.GetProducts(storage), "product", "GET /api/product"))
	router.HandleFunc("POST /api/product/quantity", pt.Instrument(handler.UpdateProductQuantity(ch, q), "product", "POST /api/product/quantity"))

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
