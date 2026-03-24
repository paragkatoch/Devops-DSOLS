package rabbitmq

import (
	"context"
	"log"
	"time"

	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
	"github.com/rabbitmq/amqp091-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

func New() (*amqp.Connection, *amqp.Channel) {
	// connect
	conn, err := amqp091.Dial("amqp://guest:guest@localhost:5672/")
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

func SendMessage(ch *amqp.Channel, q amqp.Queue, body []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	errhandler.FailOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s\n", body)
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

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
