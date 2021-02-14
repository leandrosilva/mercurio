package main

// Event is something worth enough to notify somebody else
type Event struct {
	ID            string `json:"id,omitempty"`
	SourceID      string `json:"sourceID,omitempty"`
	DestinationID string `json:"destinationID,omitempty"`
	Data          string `json:"data,omitempty"`
}

// BrodcastEvent is something worth enough to notify many targets
type BrodcastEvent struct {
	ID           string   `json:"id,omitempty"`
	SourceID     string   `json:"sourceID,omitempty"`
	Destinations []string `json:"destinations,omitempty"`
	Data         string   `json:"data,omitempty"`
}

// Client is the target notification entity
type Client struct {
	ID      string
	Channel chan Event
}
