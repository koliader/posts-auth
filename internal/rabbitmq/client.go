package rabbitmq

import (
	"fmt"

	"github.com/koliader/posts-auth.git/internal/util"
	"github.com/streadway/amqp"
)

type UpdateEmailMessage struct {
	Email    string `json:"email"`
	NewEmail string `json:"newEmail"`
}

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewClient(config util.Config, channelName string) (*Client, error) {
	conn, err := amqp.Dial(config.RbmUrl)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	_, err = channel.QueueDeclare(channelName, false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:    conn,
		channel: channel,
	}, nil
}
func (c *Client) SendMessage(queueName string, message []byte) error {
	err := c.channel.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
		},
	)
	if err != nil {
		return fmt.Errorf("error publishing message to RabbitMQ: %w", err)
	}
	return nil
}

func (c *Client) GetMessage(queueName string) (<-chan amqp.Delivery, error) {
	msgs, err := c.channel.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func (c *Client) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			return err
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}
