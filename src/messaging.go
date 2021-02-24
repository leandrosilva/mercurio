package main

// MessageQueueSettings houlds in parameters to create instances of the given message-oriented
// middleware channel
type MessageQueueSettings struct {
	Use        bool
	URL        string
	Topic      string
	RoutingKey string
}

// MessageQueueConnection is the interface to the message-oriented middleware which supports scale
// out this notification service, e.g. RabbitMQ, ActiveMQ, Amazon SQS, Redis Pub/Sub, etc
type MessageQueueConnection interface {
	Close()
	PublishNotification(notification Notification) error
	ConsumeNotifications() (MessageConsumer, error)
}

// MessageConsumer is the interface to start receive messages from the message-oriented middleware
type MessageConsumer interface {
	IsReady() bool
}
