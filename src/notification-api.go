package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// NotificationAPI is the public HTTP interface for the Broker
type NotificationAPI struct {
	Broker     *Broker
	Repository NotificationRepository
}

// NewNotificationAPI creates an instance of the NotificationAPI
func NewNotificationAPI(broker *Broker, repository NotificationRepository) (api NotificationAPI) {
	api = NotificationAPI{
		Broker:     broker,
		Repository: repository,
	}
	return
}

type notifyEventResponse struct {
	NotificationID uint   `json:"notificationID"`
	EventID        string `json:"eventID"`
}

// NotifyEventHandler is the endpoint to publishs events from source to destination
func (api *NotificationAPI) NotifyEventHandler(w http.ResponseWriter, r *http.Request) {
	var event Event
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Receiving event for client %s from source %s", event.DestinationID, event.SourceID)

	notification, err := api.Broker.NotifyEvent(event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := notifyEventResponse{
		NotificationID: notification.ID,
		EventID:        notification.EventID,
	}
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type broadEventResponse struct {
	Notifications []uint   `json:"notificationID"`
	Events        []string `json:"eventID"`
}

// BrodcastEventHandler is the endpoint to publishs events from source to many destinations
func (api *NotificationAPI) BrodcastEventHandler(w http.ResponseWriter, r *http.Request) {
	var brodcastEvent BroadcastEvent
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&brodcastEvent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Receiving event to broadcast from source %s to %s destinations", brodcastEvent.SourceID, brodcastEvent.Destinations)

	notifications, err := api.Broker.BroadcastEvent(brodcastEvent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := broadEventResponse{}
	for _, notification := range notifications {
		response.Notifications = append(response.Notifications, notification.ID)
		response.Events = append(response.Events, notification.EventID)
	}

	if err := json.NewEncoder(w).Encode(&response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type streamNotificationsResponse struct {
	NotificationID uint   `json:"notificationID,omitempty"`
	EventID        string `json:"eventID,omitempty"`
	SourceID       string `json:"sourceID,omitempty"`
	ClientID       string `json:"clientID,omitempty"`
	Data           string `json:"data,omitempty"`
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
	clientChan := make(chan Notification)

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
		// Get event for client
		notification := <-clientChan

		// Encode the event
		response := streamNotificationsResponse{
			NotificationID: notification.ID,
			EventID:        notification.EventID,
			SourceID:       notification.SourceID,
			ClientID:       notification.DestinationID,
			Data:           notification.Data,
		}
		jsonResponse, err := json.Marshal(&response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send it
		fmt.Fprintf(w, "data: %s\n\n", string(jsonResponse))

		// Flush the data immediatly instead of buffering it for later
		// so client receives it right on
		flusher.Flush()
	}
}

type notificationsResponse struct {
	ClientID      string                 `json:"clientID,omitempty"`
	Notifications []notificationResponse `json:"events"`
}

// GetNotificationsHandler responds with notifications owned by a given client
func (api *NotificationAPI) GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["clientID"]

	log.Printf("Getting notifications of client %s", clientID)

	w.Header().Set("Content-Type", "application/json")

	notifications, err := api.Repository.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := notificationsResponse{
		ClientID:      clientID,
		Notifications: []notificationResponse{},
	}
	for _, notification := range notifications {
		response.Notifications = append(response.Notifications, notificationResponse{
			NotificationID: notification.ID,
			EventID:        notification.EventID,
			SourceID:       notification.SourceID,
			Data:           notification.Data,
			CreatedAt:      notification.CreatedAt,
			ReadAt:         notification.ReadAt,
		})
	}

	if err := json.NewEncoder(w).Encode(&response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type notificationResponse struct {
	NotificationID uint       `json:"notificationID,omitempty"`
	EventID        string     `json:"eventID,omitempty"`
	SourceID       string     `json:"sourceID,omitempty"`
	ClientID       string     `json:"clientID,omitempty"`
	Data           string     `json:"data,omitempty"`
	CreatedAt      time.Time  `json:"createdAt,omitempty"`
	ReadAt         *time.Time `json:"readAt,omitempty"`
}

// GetNotificationHandler responds with a event notification by its id
func (api *NotificationAPI) GetNotificationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["clientID"]
	notificationID, err := strconv.Atoi(vars["notificationID"])
	if err != nil {
		http.Error(w, "notificationID not in a valid format", http.StatusBadRequest)
		return
	}

	log.Printf("Getting notification %d of client %s", notificationID, clientID)

	w.Header().Set("Content-Type", "application/json")

	notification, err := api.Repository.Get(uint(notificationID))
	if err != nil {
		if errors.Is(err, ErrNotificationNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := notificationResponse{
		NotificationID: notification.ID,
		EventID:        notification.EventID,
		SourceID:       notification.SourceID,
		Data:           notification.Data,
		CreatedAt:      notification.CreatedAt,
		ReadAt:         notification.ReadAt,
	}

	if err := json.NewEncoder(w).Encode(&response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type changeNotificationStatusResponse struct {
	Status string `json:"status,omitempty"`
}

// MarkNotificationReadHandler changes the notificatinn status to read
func (api *NotificationAPI) MarkNotificationReadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["clientID"]
	notificationID, err := strconv.Atoi(vars["notificationID"])
	if err != nil {
		http.Error(w, "notificationID not in a valid format", http.StatusBadRequest)
		return
	}

	notification, err := api.Repository.Get(uint(notificationID))
	if err != nil {
		if errors.Is(err, ErrNotificationNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// A read notification is simply one that has a read time
	readAt := time.Now()
	notification.ReadAt = &readAt

	err = api.Repository.Update(&notification)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Marking notification %d of client %s as read", notificationID, clientID)

	w.Header().Set("Content-Type", "application/json")

	response := changeNotificationStatusResponse{
		Status: "read",
	}
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// MarkNotificationUnreadHandler changes the notificatinn status to unread
func (api *NotificationAPI) MarkNotificationUnreadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["clientID"]
	notificationID, err := strconv.Atoi(vars["notificationID"])
	if err != nil {
		http.Error(w, "notificationID not in a valid format", http.StatusBadRequest)
		return
	}

	notification, err := api.Repository.Get(uint(notificationID))
	if err != nil {
		if errors.Is(err, ErrNotificationNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// A read notification is simply one that does not have a read time
	notification.ReadAt = nil

	err = api.Repository.Update(&notification)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Marking notification %d of client %s as unread", notificationID, clientID)

	w.Header().Set("Content-Type", "application/json")

	response := changeNotificationStatusResponse{
		Status: "unread",
	}
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
