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

// RabbitMQConnection is a wrapper to an open connection to a RabbitMQ server, with two-way channel (i.e. pub/sub, producer/consumer)
type RabbitMQConnection struct {
	// The service node ID where this broken is running in
	nid string

	connection *amqp.Connection
	pubChannel *amqp.Channel
	subChannel *amqp.Channel
	topic      string
	routingKey string
	queue      string
}

// NewRabbitMQConnection opens a new TCP connection to a RabbitMQ server and gets pub/sub channels setted up too
func NewRabbitMQConnection(nid string, settings MessageQueueSettings) (*RabbitMQConnection, error) {
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

	rabbit := &RabbitMQConnection{
		nid:        nid,
		connection: conn,
		pubChannel: pch,
		subChannel: sch,
		topic:      settings.Topic,
		routingKey: settings.RoutingKey,
	}

	log.Printf("Connected to RabbitMQ at '%s'", settings.URL)

	return rabbit, nil
}

// Close underlying connection and channels
func (mq *RabbitMQConnection) Close() {
	mq.subChannel.Close()
	mq.pubChannel.Close()
	mq.connection.Close()
}

// PublishNotification send a notification to a RabbitMQ topic with the given routing key
func (mq *RabbitMQConnection) PublishNotification(notification Notification) error {
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
			AppId:       mq.nid,
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
func (mq *RabbitMQConnection) ConsumeNotifications() (MessageConsumer, error) {
	msgs, err := mq.subChannel.Consume(
		mq.queue, // queue
		"",       // consumer
		true,     // auto ack
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
