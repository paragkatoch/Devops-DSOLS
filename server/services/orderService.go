package services

import (
	"encoding/json"
	"log"
	"log/slog"

	config "github.com/paragkatoch/Devops-DSOLS/internal"
	"github.com/paragkatoch/Devops-DSOLS/internal/rabbitmq"
	"github.com/paragkatoch/Devops-DSOLS/internal/storage"
	"github.com/paragkatoch/Devops-DSOLS/types"
)

func OrderService(storage storage.Storage, cfg *config.Config) {
	slog.Info("Hello from order service")

	// connect to queue
	conn, ch := rabbitmq.New(cfg)

	defer conn.Close()
	defer ch.Close()

	q := rabbitmq.Connect(ch, "order")

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
