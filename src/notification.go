package main

// Event is something worth enough to notify somebody else
type Event struct {
	ID            string
	SourceID      string
	DestinationID string
	Data          string
}

// BrodcastEvent is something worth enough to notify many targets
type BrodcastEvent struct {
	ID                string
	SourceID          string
	DestinationListID []string
	Data              string
}

// Client is the target notification entity
type Client struct {
	ID      string
	Channel chan Event
}
