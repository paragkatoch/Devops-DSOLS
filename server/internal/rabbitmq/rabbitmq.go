package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"time"

	config "github.com/paragkatoch/Devops-DSOLS/internal"
	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
	serverhandler "github.com/paragkatoch/Devops-DSOLS/util/serverHandler"
	amqp "github.com/rabbitmq/amqp091-go"
)

func New(cfg *config.Config) (*amqp.Connection, *amqp.Channel) {
	// connect
	var conn *amqp.Connection
	var err error

	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(cfg.Queue_path)
		if err == nil {
			break
		}

		log.Println("RabbitMQ not ready, retrying...")
		time.Sleep(2 * time.Second)
	}
	// conn, err = amqp.Dial(cfg.Queue_path)
	errhandler.FailOnError(err, "Failed to connect to RabbitMQ")

	// get channel
	ch, err := conn.Channel()
	errhandler.FailOnError(err, "Failed to open a channel")

	return conn, ch
}

func Connect(ch *amqp.Channel, channel string) amqp.Queue {
	q, err := ch.QueueDeclare(
		channel, // name
		true,    // durability
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		amqp.Table{
			amqp.QueueTypeArg: amqp.QueueTypeQuorum,
		},
	)
	errhandler.FailOnError(err, "Failed to declare a queue")
	return q
}

func SendMessage(ch *amqp.Channel, q amqp.Queue, body interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonBody,
		})

	if err != nil {
		return err
	}

	log.Println(" [+] Sent a message")
	return nil
}

func ReceiveMessage(ch *amqp.Channel, q amqp.Queue, handler func([]byte)) {
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	errhandler.FailOnError(err, "Failed to register a consumer")

	serverhandler.Async(func() {
		log.Println(" [*] Waiting for messages.")

		for d := range msgs {
			log.Println(" [-] Received a message")
			handler(d.Body)
		}
	})
}
