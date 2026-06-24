package amqp

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

var MainAMQPClient *AMQPClient

// TODO: ensure it is thread-safe
type AMQPClient struct {
	conn *amqp091.Connection
	ch   *amqp091.Channel
}

func NewAMQPClient() (*AMQPClient, error) {
	client := &AMQPClient{}

	amqpURL, found := os.LookupEnv("AMQP_URL")
	if !found {
		return nil, fmt.Errorf("AMQP_URL not found in env")
	}
	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("connecting to AMQP: %s", err)
	}
	client.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("creating a channel: %s", err)
	}
	client.ch = ch

	return client, nil
}

func (client *AMQPClient) Close() error {
	// TODO: do we have to deal with the errors?
	client.ch.Close()
	return client.conn.Close()
}

// TODO: do not create the queue for every request!
// TODO: accept []byte instead of string as a message ??
func (client *AMQPClient) SendMessageToQueue(message string, queueName string) error {
	q, err := client.ch.QueueDeclare(
		queueName, true, false, false, false, amqp091.Table{
			amqp091.QueueTypeArg: amqp091.QueueTypeQuorum,
		})
	if err != nil {
		return fmt.Errorf("declaring a queue: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.ch.PublishWithContext(
		ctx, "", q.Name, false, false, amqp091.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	if err != nil {
		return fmt.Errorf("sending the message: %s", err)
	}

	slog.Debug("sent AMQP message", "message", message, "queue", queueName)

	return nil
}
