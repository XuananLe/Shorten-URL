package stores

import (
    amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConfig struct {
    Conn    *amqp.Connection
    Channel *amqp.Channel
}

func NewRabbitMQConfig() (*RabbitMQConfig, error) {
    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
    if err != nil {
        return nil, err
    }

    channel, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, err
    }

    return &RabbitMQConfig{Conn: conn, Channel: channel}, nil
}
