package stores

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Queues  map[string]amqp.Queue
}

var (
	RabbitMQClient *RabbitMQ
)

func InitRabbitMQ() *RabbitMQ {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		log.Fatalf("Failed to open a channel: %v", err)
	}

	fmt.Println("RabbitMQ Connected")

	RabbitMQClient = &RabbitMQ{
		Conn:    conn,
		Channel: channel,
		Queues:  make(map[string]amqp.Queue),
	}

	return RabbitMQClient
}

func (r *RabbitMQ) DeclareQueue(queueName string) (amqp.Queue, error) {
	queue, err := r.Channel.QueueDeclare(
		queueName, 
		true,      
		false,     
		false,     
		false,     
		nil,      
	)
	if err != nil {
		return amqp.Queue{}, err
	}
	fmt.Println("Created queue", queueName)

	r.Queues[queueName] = queue
	return queue, nil
}

func CloseRabbitMQ() {
	if RabbitMQClient != nil {
		RabbitMQClient.Channel.Close()
		RabbitMQClient.Conn.Close()
	}
}
