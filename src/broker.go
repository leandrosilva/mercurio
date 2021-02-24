package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// Broker is the core notification service entity
type Broker struct {
	// The service node ID where this broken is running in
	nid string

	// Flags whether broker is isRunning or not
	isRunning bool

	// The underlying datastore for notifications persistence
	repository NotificationRepository

	// The underlying message-orinted middleware (might be nil if it does not uses one; it depends on settings passed by on creation)
	mq MessageQueueConnection

	// Events are pushed to this channel by the main events-gathering routine
	notifications chan Notification

	// New client connections
	newClients chan Client

	// Closed client connections
	closingClients chan Client

	// Client connections registry
	clients map[string]Client
}

// NewBroker creates a new Broker and puts it to run
func NewBroker(nid string, repository NotificationRepository, mqSettings MessageQueueSettings) (*Broker, error) {
	broker := &Broker{
		nid:            nid,
		isRunning:      false,
		repository:     repository,
		notifications:  make(chan Notification, 1),
		newClients:     make(chan Client),
		closingClients: make(chan Client),
		clients:        make(map[string]Client),
	}

	// We're assuming RabbitMQ here but we can change it in the future and encapsulate it another way
	// in a factory or something
	if mqSettings.Use {
		mq, err := NewRabbitMQConnection(nid, mqSettings)
		if err != nil {
			return nil, fmt.Errorf("failed to connect & setup a channel with RabbitMQ server due to: %s", err)
		}
		broker.mq = mq
	}

	return broker, nil
}

// Run starts of the Broker notification service
func (b *Broker) Run() error {
	b.isRunning = true

	// As we know we're working with RabbitMQ in the current incarnation of Mercurio, let't make thing
	// a bit specific here
	var incomeMessages <-chan amqp.Delivery
	if b.mq != nil {
		consummer, err := b.mq.ConsumeNotifications()
		if err != nil {
			b.isRunning = false
			return err
		}
		incomeMessages = consummer.(*RabbitMQConsumer).IncomeMessages
	}

	// The message exchange goroutine
	go func() {
		for b.isRunning {
			select {
			case c := <-b.newClients:
				// A new client has connected
				// Register their message channel
				b.clients[c.ID] = c
				log.Printf("Client added. (%d registered clients)", len(b.clients))

			case c := <-b.closingClients:
				// A client has dettached and we want to
				// stop sending them messages.
				delete(b.clients, c.ID)
				log.Printf("Removed client. (%d registered clients)", len(b.clients))

			case notification := <-b.notifications:
				// We got a new event from the outside!
				// Should notify the destination client
				clientID := notification.DestinationID
				client, exists := b.clients[notification.DestinationID]

				log.Printf("Got notification %d for client %s (known = %v)", notification.ID, clientID, exists)

				if exists {
					client.Channel <- notification
					log.Printf("Send notification %d to client %s", notification.ID, clientID)
				}

				if b.mq != nil {
					// Publish message to MQ -- maybe should have an additional condition here to decide
					// whether or to send the notification to MQ
					if !exists {
						log.Printf("Publish to MQ notification %d for client %s. (which is unknown to this service node)", notification.ID, clientID)
						b.mq.PublishNotification(notification)
					}
				}

			// It receives messages that was published by itself, auto acks it (its queue is exclusive) and move forward without do anything else;
			// otherwise when receiving messages published by other service nodes, if the client is known here, pushes the notification as normal
			case message := <-incomeMessages:
				if len(message.Body) == 0 {
					continue
				}

				if message.AppId == b.nid {
					log.Printf("Message %s was published by myself, skipping it...", message.MessageId)
					continue
				}

				notification, err := UnmarshalNotification(message.Body)
				if err != nil {
					log.Printf("Could not unmarshal message body due to: %s", err)
				}

				clientID := notification.DestinationID
				client, exists := b.clients[notification.DestinationID]

				log.Printf("Got from MQ with notification %d for client %s (known = %v)", notification.ID, clientID, exists)

				if exists {
					client.Channel <- notification
					log.Printf("Send notification %d got from MQ to client %s", notification.ID, clientID)
				}
			default:
			}
		}
	}()

	return nil
}

// Stop shuts down the Broker notification service
func (b *Broker) Stop() {
	b.isRunning = false

	if b.mq != nil {
		log.Println("Closing MQ channel")
		b.mq.Close()
	}
}

// NotifyClientConnected notifies a new client has arrived
func (b *Broker) NotifyClientConnected(client Client) {
	b.newClients <- client
}

// NotifyClientDisconnected notifies a client is gone
func (b *Broker) NotifyClientDisconnected(client Client) {
	b.closingClients <- client
}

// NotifyEvent when an event has occourred for one destination
func (b *Broker) NotifyEvent(event Event) (Notification, error) {
	notification, err := NewNotification(&event)
	if err != nil {
		return Notification{}, err
	}

	err = b.repository.Add(notification)
	if err != nil {
		return Notification{}, err
	}

	b.notifications <- *notification

	return *notification, nil
}

// BroadcastEvent when an event has occourred for many destinations
func (b *Broker) BroadcastEvent(broadcastEvent BroadcastEvent) ([]Notification, error) {
	notifications := []Notification{}

	for _, destinationID := range broadcastEvent.Destinations {
		event := Event{
			ID:            broadcastEvent.ID,
			SourceID:      broadcastEvent.SourceID,
			DestinationID: destinationID,
			Data:          broadcastEvent.Data,
		}

		notification, err := b.NotifyEvent(event)
		if err != nil {
			return nil, err
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}
