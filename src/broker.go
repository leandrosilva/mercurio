package main

import (
	"fmt"
	"log"
)

// Broker is the core notification service entity
type Broker struct {
	// Flags whether broker is isRunning or not
	isRunning bool

	// The underlying datastore for notifications persistence
	repository NotificationRepository

	// The underlying message-orinted middleware
	mq MessageQueueChannel

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
func NewBroker(repository NotificationRepository, mqSettings MessageQueueSettings) (*Broker, error) {
	broker := &Broker{
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
		mq, err := NewRabbitMQChannel(mqSettings)
		if err != nil {
			return nil, fmt.Errorf("failed to connect & setup a channel with RabbitMQ server due to: %s", err)
		}
		broker.mq = mq
	}

	return broker, nil
}

// Run starts of the Broker notification service
func (broker *Broker) Run() error {
	broker.isRunning = true

	// As we know we're working with RabbitMQ in the current incarnation of Mercurio, let't make thing
	// a bit specific here
	var rabbitMQ *RabbitMQConsumer
	if broker.mq != nil {
		consummer, err := broker.mq.ConsumeNotifications()
		if err != nil {
			broker.isRunning = false
			return err
		}
		rabbitMQ = consummer.(*RabbitMQConsumer)
	}

	// The message exchange goroutine
	go func() {
		for broker.isRunning {
			select {
			case c := <-broker.newClients:
				// A new client has connected
				// Register their message channel
				broker.clients[c.ID] = c
				log.Printf("Client added. (%d registered clients)", len(broker.clients))

			case c := <-broker.closingClients:
				// A client has dettached and we want to
				// stop sending them messages.
				delete(broker.clients, c.ID)
				log.Printf("Removed client. (%d registered clients)", len(broker.clients))

			case notification := <-broker.notifications:
				// We got a new event from the outside!
				// Should notify the destination client
				clientID := notification.DestinationID
				client, exists := broker.clients[notification.DestinationID]
				if exists {
					client.Channel <- notification
					log.Printf("Send notification %d to client %s", notification.ID, clientID)
				}

				if broker.mq != nil {
					// Publish message to MQ -- maybe should have an additional condition here to decide
					// whether or to send the notification to MQ
					if !exists {
						log.Printf("Publish to MQ notification %d for client %s. (not in this service node)", notification.ID, clientID)
						broker.mq.PublishNotification(notification)
					}
				}
			default:
			}

			if broker.mq != nil {
				select {
				case message := <-rabbitMQ.IncomeMessages:
					if len(message.Body) == 0 {
						continue
					}

					notification, err := UnmarshalNotification(message.Body)
					if err != nil {
						log.Printf("Could not unmarshal message body due to: %s", err)
					}

					clientID := notification.DestinationID
					log.Printf("Got from MQ with notification %d for client %s", notification.ID, clientID)

					client, exists := broker.clients[notification.DestinationID]
					if exists {
						client.Channel <- notification
						log.Printf("Send notification %d got from MQ to client %s", notification.ID, clientID)
					}
				default:
				}
			}
		}
	}()

	return nil
}

// Stop shuts down the Broker notification service
func (broker *Broker) Stop() {
	broker.isRunning = false

	if broker.mq != nil {
		log.Println("Closing MQ channel")
		broker.mq.CloseChannel()
	}
}

// NotifyClientConnected notifies a new client has arrived
func (broker *Broker) NotifyClientConnected(client Client) {
	broker.newClients <- client
}

// NotifyClientDisconnected notifies a client is gone
func (broker *Broker) NotifyClientDisconnected(client Client) {
	broker.closingClients <- client
}

// NotifyEvent when an event has occourred for one destination
func (broker *Broker) NotifyEvent(event Event) (Notification, error) {
	notification, err := NewNotification(&event)
	if err != nil {
		return Notification{}, err
	}

	err = broker.repository.Add(notification)
	if err != nil {
		return Notification{}, err
	}

	broker.notifications <- *notification

	return *notification, nil
}

// BroadcastEvent when an event has occourred for many destinations
func (broker *Broker) BroadcastEvent(broadcastEvent BroadcastEvent) ([]Notification, error) {
	notifications := []Notification{}

	for _, destinationID := range broadcastEvent.Destinations {
		event := Event{
			ID:            broadcastEvent.ID,
			SourceID:      broadcastEvent.SourceID,
			DestinationID: destinationID,
			Data:          broadcastEvent.Data,
		}

		notification, err := broker.NotifyEvent(event)
		if err != nil {
			return nil, err
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}
