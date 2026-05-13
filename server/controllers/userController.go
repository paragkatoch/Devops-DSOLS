package controllers

import (
	"log/slog"
	"net/http"

	config "github.com/paragkatoch/Devops-DSOLS/internal"
	handler "github.com/paragkatoch/Devops-DSOLS/internal/http/handlers/user"
	"github.com/paragkatoch/Devops-DSOLS/internal/prometheus"
	"github.com/paragkatoch/Devops-DSOLS/internal/rabbitmq"
	"github.com/paragkatoch/Devops-DSOLS/internal/storage"
	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
	serverhandler "github.com/paragkatoch/Devops-DSOLS/util/serverHandler"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func UserController(storage storage.Storage, cfg *config.Config) {
	slog.Info("Hello from user controller")

	// connect to queue
	conn, ch := rabbitmq.New(cfg)

	defer conn.Close()
	defer ch.Close()

	q := rabbitmq.Connect(ch, "user")
	publish := rabbitmq.GetPublisher(conn, q)

	// setup router
	router := http.NewServeMux()

	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("POST /api/user", prometheus.Instrument(handler.CreateUser(ch, publish), "user", "post_user"))
	// router.HandleFunc("GET /api/user/{id}/order", prometheus.Instrument(handler.GetUserOrders(storage), "user", "get_order"))
	router.HandleFunc("GET /api/user/{id}", prometheus.Instrument(handler.GetUser(storage), "user", "get_user"))
	router.HandleFunc("GET /api/user", prometheus.Instrument(handler.GetUsers(storage), "user", "get_users"))

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
