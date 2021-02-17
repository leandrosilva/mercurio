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

type unicastEventResponse struct {
	NotificationID uint   `json:"notificationID"`
	EventID        string `json:"eventID"`
}

// UnicastEventHandler is the endpoint to publishs events from one source to one destination
func (api *NotificationAPI) UnicastEventHandler(w http.ResponseWriter, r *http.Request) {
	var event Event
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&event)
	if err != nil {
		respondWithBadRequest(w, err.Error())
		return
	}

	log.Printf("Receiving event for client %s from source %s", event.DestinationID, event.SourceID)

	notification, err := api.Broker.NotifyEvent(event)
	if err != nil {
		respondWithInternalServerError(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := unicastEventResponse{
		NotificationID: notification.ID,
		EventID:        notification.EventID,
	}
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		respondWithInternalServerError(w, err.Error())
	}
}

type broadcastEventResponse struct {
	NotificationID uint   `json:"notificationID"`
	EventID        string `json:"eventID"`
}

// BroadcastEventHandler is the endpoint to publishs events from one source to many destinations
func (api *NotificationAPI) BroadcastEventHandler(w http.ResponseWriter, r *http.Request) {
	var brodcastEvent BroadcastEvent
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&brodcastEvent)
	if err != nil {
		respondWithBadRequest(w, err.Error())
		return
	}

	log.Printf("Receiving event to broadcast from source %s to %s destinations", brodcastEvent.SourceID, brodcastEvent.Destinations)

	notifications, err := api.Broker.BroadcastEvent(brodcastEvent)
	if err != nil {
		respondWithInternalServerError(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := []broadcastEventResponse{}
	for _, notification := range notifications {
		response = append(response, broadcastEventResponse{
			NotificationID: notification.ID,
			EventID:        notification.EventID,
		})
	}

	if err := json.NewEncoder(w).Encode(&response); err != nil {
		respondWithInternalServerError(w, err.Error())
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
		respondWithInternalServerError(w, "Streaming is unsupported.")
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
			respondWithInternalServerError(w, err.Error())
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
	Notifications []notificationResponse `json:"notifications"`
}

// GetNotificationsHandler responds with notifications owned by a given client
func (api *NotificationAPI) GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["clientID"]

	// Optional query strings
	status := r.FormValue("status")
	if !IsValidNotificationStatus(status) {
		respondWithBadRequest(w, fmt.Sprintf("%s is not a valid status", status))
		return
	}

	log.Printf("Getting notifications of client %s", clientID)

	notifications, err := api.Repository.GetByStatus(clientID, status)
	if err != nil {
		respondWithInternalServerError(w, err.Error())
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

	respondWithSuccess(w, response)
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
	notificationID, _ := strconv.Atoi(vars["notificationID"])

	log.Printf("Getting notification %d of client %s", notificationID, clientID)

	notification, err := api.Repository.Get(clientID, uint(notificationID))
	if err != nil {
		if errors.Is(err, ErrNotificationNotFound) {
			respondWithNotFound(w, err.Error())
			return
		}
		respondWithInternalServerError(w, err.Error())
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

	respondWithSuccess(w, response)
}

type changeNotificationStatusResponse struct {
	Status string `json:"status,omitempty"`
}

// MarkNotificationReadHandler changes the notificatinn status to read
func (api *NotificationAPI) MarkNotificationReadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["clientID"]
	notificationID, _ := strconv.Atoi(vars["notificationID"])

	notification, err := api.Repository.Get(clientID, uint(notificationID))
	if err != nil {
		if errors.Is(err, ErrNotificationNotFound) {
			respondWithNotFound(w, err.Error())
			return
		}
		respondWithInternalServerError(w, err.Error())
		return
	}

	// A read notification is simply one that has a read time
	readAt := time.Now()
	notification.ReadAt = &readAt

	err = api.Repository.Update(&notification)
	if err != nil {
		respondWithInternalServerError(w, err.Error())
		return
	}

	log.Printf("Marking notification %d of client %s as read", notificationID, clientID)

	response := changeNotificationStatusResponse{
		Status: "read",
	}

	respondWithSuccess(w, response)
}

// MarkNotificationUnreadHandler changes the notificatinn status to unread
func (api *NotificationAPI) MarkNotificationUnreadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["clientID"]
	notificationID, _ := strconv.Atoi(vars["notificationID"])

	notification, err := api.Repository.Get(clientID, uint(notificationID))
	if err != nil {
		if errors.Is(err, ErrNotificationNotFound) {
			respondWithNotFound(w, err.Error())
			return
		}
		respondWithInternalServerError(w, err.Error())
		return
	}

	// A read notification is simply one that does not have a read time
	notification.ReadAt = nil

	err = api.Repository.Update(&notification)
	if err != nil {
		respondWithInternalServerError(w, err.Error())
		return
	}

	log.Printf("Marking notification %d of client %s as unread", notificationID, clientID)

	response := changeNotificationStatusResponse{
		Status: "unread",
	}

	respondWithSuccess(w, response)
}
