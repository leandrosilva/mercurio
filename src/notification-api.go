package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// NotificationAPI is the public HTTP interface for the Broker
type NotificationAPI struct {
	Broker Broker
}

// NewNotificationAPI creates an instance of the NotificationAPI
func NewNotificationAPI(broker *Broker) (api NotificationAPI) {
	api = NotificationAPI{
		Broker: *broker,
	}
	return
}

// StreamNotificationsHandler is the endpoint for clients listening for notifications
func (api *NotificationAPI) StreamNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if SSE is possible
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming is unsupported.", http.StatusInternalServerError)
		return
	}

	// SSE support headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Registers client connection with the Broker
	vars := mux.Vars(r)
	clientID := vars["clientID"]
	clientChan := make(chan Event)

	client := Client{
		ID:      clientID,
		Channel: clientChan,
	}

	api.Broker.NotifyClientConnected(client)

	// Remove this client from the map of connected clients when this handler exits
	defer func() {
		api.Broker.NotifyClientDisconnected(client)
	}()

	// Unregisters closed connections
	notifyClosed := r.Context().Done()
	go func() {
		<-notifyClosed
		api.Broker.NotifyClientDisconnected(client)
	}()

	for {
		// Send events to client
		event := <-clientChan
		fmt.Fprintf(w,
			"data: {\"id\":\"%s\",\"sourceID\":\"%s\",\"clientID\":\"%s\",\"content\":\"%s\"}\n\n",
			event.ID, event.SourceID, event.ClientID, event.Data)

		// Flush the data immediatly instead of buffering it for later
		// so client receives it right on
		flusher.Flush()
	}
}

// NotifyEventHandler is the endpoint to publishs events fromt source to client
func (api *NotificationAPI) NotifyEventHandler(w http.ResponseWriter, r *http.Request) {
	var event Event
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	event.ID = fmt.Sprint(time.Now().Unix())

	log.Printf("Receiving event for client %s from source %s", event.ClientID, event.SourceID)
	api.Broker.NotifyEvent(event)

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, "{\"eventID\":\"%s\"}\n", event.ID)
}

// GetNotificationsHandler responds with notifications owned by a given client
func (api *NotificationAPI) GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["clientID"]

	log.Printf("Getting notificatins of client %s", clientID)

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, "{\"clientID\":\"%s\",\"notifications\":[]}\n", clientID)
}
