package main

// Event is something worth enough to notify somebody else
type Event struct {
	ID       string
	SourceID string
	ClientID string
	Data     string
}
