package main

import (
	"log"
)

// Broker is the core notification service entity
type Broker struct {
	// Events are pushed to this channel by the main events-gathering routine
	notifications chan Event

	// New client connections
	newClients chan Client

	// Closed client connections
	closingClients chan Client

	// Client connections registry
	clients map[string]Client
}

// NewBroker creates a new Broker and puts it to run
func NewBroker() (broker *Broker) {
	broker = &Broker{
		notifications:  make(chan Event, 1),
		newClients:     make(chan Client),
		closingClients: make(chan Client),
		clients:        make(map[string]Client),
	}
	return
}

// Run starts of the Broker notification service
func (broker *Broker) Run() {
	go func() {
		for {
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

			case event := <-broker.notifications:
				// We got a new event from the outside!
				// Send event to the target client
				client, exists := broker.clients[event.ClientID]
				if exists {
					client.Channel <- event
					log.Printf("Send event to client %s ", client.ID)
				}
			}
		}
	}()
}

// NotifyClientConnected notifies a new client has arrived
func (broker *Broker) NotifyClientConnected(client Client) {
	broker.newClients <- client
}

// NotifyClientDisconnected notifies a client is gone
func (broker *Broker) NotifyClientDisconnected(client Client) {
	broker.closingClients <- client
}

// NotifyEvent when an event has occourred
func (broker *Broker) NotifyEvent(event Event) {
	// TODO: save event before notify it has happened
	broker.notifications <- event
}
