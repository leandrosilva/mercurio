package main

import (
	"time"

	"github.com/google/uuid"
)

// Notification is the persistent record of a known event (read/unread)
type Notification struct {
	ID            uint      `json:"id,omitempty" gorm:"primaryKey"`
	EventID       string    `json:"event,omitempty" gorm:"index"`
	SourceID      string    `json:"sourceID,omitempty" gorm:"index"`
	DestinationID string    `json:"destinationID,omitempty" gorm:"index"`
	Data          string    `json:"data,omitempty"`
	CreatedAt     time.Time `json:"createdAt,omitempty"`
	ReadAt        time.Time `json:"readAt,omitempty"`
}

// NewNotification creates a new notification for a given event
func NewNotification(event *Event) (*Notification, error) {
	// In case event doesn't already have an ID, give it a unique one
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	notification := &Notification{
		EventID:       event.ID,
		SourceID:      event.SourceID,
		DestinationID: event.DestinationID,
		Data:          event.Data,
	}

	return notification, nil
}

// Event is something worth enough to be notified
type Event struct {
	ID            string `json:"id,omitempty"`
	SourceID      string `json:"sourceID,omitempty"`
	DestinationID string `json:"destinationID,omitempty"`
	Data          string `json:"data,omitempty"`
}

// BroadcastEvent is something worth enough to be broadcasted
type BroadcastEvent struct {
	ID           string   `json:"id,omitempty"`
	SourceID     string   `json:"sourceID,omitempty"`
	Destinations []string `json:"destinations,omitempty"`
	Data         string   `json:"data,omitempty"`
}

// Client is the target notification entity
type Client struct {
	ID      string
	Channel chan Notification
}
