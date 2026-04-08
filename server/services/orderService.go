package services

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"

	config "github.com/paragkatoch/Devops-DSOLS/internal"
	"github.com/paragkatoch/Devops-DSOLS/internal/rabbitmq"
	"github.com/paragkatoch/Devops-DSOLS/internal/storage"
	"github.com/paragkatoch/Devops-DSOLS/types"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func OrderService(storage storage.Storage, cfg *config.Config) {
	slog.Info("Hello from order service")

	// connect to queue
	conn, ch := rabbitmq.New(cfg)

	defer conn.Close()
	defer ch.Close()

	q := rabbitmq.Connect(ch, "order")

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(cfg.HTTPServer.Addr, nil)
	}()

	// receive messages from queue
	rabbitmq.ReceiveMessage(ch, q, func(b []byte) {
		var order types.Order
		err := json.Unmarshal(b, &order)
		if err != nil {
			log.Println("Error parsing message: ", err.Error())
			return
		}

		log.Println("Order received:", order)
	})
}
