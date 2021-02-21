package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// RabbitMQConsumer is a wrapper over a AMQP consumer channel
type RabbitMQConsumer struct {
	IncomeMessages <-chan amqp.Delivery
}

// IsReady tells whether the AMQP channel is open or what
func (mc *RabbitMQConsumer) IsReady() bool {
	return mc.IncomeMessages != nil
}

// RabbitMQChannel is a wrapper to an open channel to a RabbitMQ server
type RabbitMQChannel struct {
	connection *amqp.Connection
	pubChannel *amqp.Channel
	subChannel *amqp.Channel
	topic      string
	routingKey string
	queue      string
}

// NewRabbitMQChannel opens a new TCP connection to a RabbitMQ server and gets channel setted up too
func NewRabbitMQChannel(settings MessageQueueSettings) (*RabbitMQChannel, error) {
	conn, err := amqp.Dial(settings.URL)
	if err != nil {
		return nil, err
	}

	pch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	sch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Setup
	//

	err = pch.ExchangeDeclare(
		settings.Topic, // name
		"topic",        // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return nil, err
	}

	q, err := sch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)

	err = sch.QueueBind(
		q.Name,              // queue name
		settings.RoutingKey, // routing key
		settings.Topic,      // exchange
		false,
		nil)
	if err != nil {
		return nil, err
	}

	rabbit := &RabbitMQChannel{
		connection: conn,
		pubChannel: pch,
		subChannel: sch,
		topic:      settings.Topic,
		routingKey: settings.RoutingKey,
	}

	log.Printf("Connected to RabbitMQ at '%s'", settings.URL)

	return rabbit, nil
}

// CloseChannel underlying channel and connection
func (mq *RabbitMQChannel) CloseChannel() {
	mq.subChannel.Close()
	mq.pubChannel.Close()
	mq.connection.Close()
}

// PublishNotification send a notification to a RabbitMQ topic with the given routing key
func (mq *RabbitMQChannel) PublishNotification(notification Notification) error {
	body, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	err = mq.pubChannel.Publish(
		mq.topic,      // exchange
		mq.routingKey, // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			MessageId:   fmt.Sprintf("%d", notification.ID),
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return err
	}

	return nil
}

// ConsumeNotifications gets a MessageConsumer for consuming messages from a RabbitMQ topic
func (mq *RabbitMQChannel) ConsumeNotifications() (MessageConsumer, error) {
	msgs, err := mq.subChannel.Consume(
		mq.queue, // queue
		"",       // consumer
		false,    // auto ack
		false,    // exclusive
		false,    // no local
		false,    // no wait
		nil,      // args
	)
	if err != nil {
		return nil, err
	}

	consummer := &RabbitMQConsumer{
		IncomeMessages: msgs,
	}

	return consummer, nil
}

// UnmarshalNotification decodes a JSON notification
func UnmarshalNotification(jsonNotification []byte) (Notification, error) {
	var notification Notification
	err := json.Unmarshal(jsonNotification, &notification)
	if err != nil {
		return Notification{}, err
	}

	return notification, nil
}
