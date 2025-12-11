package rabbitmq

import (
	"Chronos/internal/config"
	"Chronos/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"

	amqp "github.com/rabbitmq/amqp091-go"
)

const mainQueue = "mainQueue"
const mainExchange = "mainExchange"

type Broker struct {
	config config.Broker
	//consumer *rabbitmq.Consumer
	producer *rabbitmq.Publisher
	client   *rabbitmq.RabbitClient
}

func NewBroker(config config.Broker) (*Broker, error) {

	client, err := rabbitmq.NewClient(rabbitmq.ClientConfig{

		URL:            config.URL,
		ConnectionName: config.ConnectionName,
		ConnectTimeout: config.ConnectTimeout,
		Heartbeat:      config.Heartbeat,

		ConsumeRetry: retry.Strategy{
			Attempts: config.ConsumeRetryAttempts,
			Delay:    config.ConsumeRetryDelay,
			Backoff:  config.ConsumeRetryBackoff}})

	if err != nil {
		return nil, err
	}

	err = client.DeclareExchange(mainExchange, "direct", true, false, false, nil)
	if err != nil {
		client.Close()
		return nil, err
	}

	err = client.DeclareQueue(mainQueue, mainExchange, mainQueue, true, false, true, nil)
	if err != nil {
		client.Close()
		return nil, err
	}

	producer := rabbitmq.NewPublisher(client, mainExchange, "application/json")

	return &Broker{config: config, producer: producer, client: client}, nil

}

func (b *Broker) Produce(ctx context.Context, notification models.Notification) error {

	id := strconv.FormatInt(notification.ID, 10)
	sendAt := time.Until(notification.SendAt)

	return retry.DoContext(ctx, retry.Strategy{
		Attempts: b.config.PublishRetryAttempts,
		Delay:    b.config.PublishRetryDelay,
		Backoff:  b.config.PublishRetryBackoff}, func() error {

		queueArgs := amqp.Table{
			"x-message-ttl":             int64(sendAt.Milliseconds()),
			"x-dead-letter-exchange":    mainExchange,
			"x-dead-letter-routing-key": mainQueue,
			"x-expires":                 int64(sendAt.Milliseconds() + b.config.MessageQueueTTL.Milliseconds()),
		}

		err := b.client.DeclareQueue(id, mainExchange, id, false, true, true, queueArgs)
		if err != nil {
			return err
		}

		ch, err := b.client.GetChannel()
		if err != nil {
			return err
		}
		defer ch.Close()

		body, err := json.Marshal(notification)
		if err != nil {
			return fmt.Errorf("failed to marshal notification to json: %w", err)
		}

		pub := amqp.Publishing{ContentType: "application/json", Body: body}

		if err := ch.PublishWithContext(ctx, mainExchange, id, false, false, pub); err != nil {
			return err
		}

		return nil

	})

}
