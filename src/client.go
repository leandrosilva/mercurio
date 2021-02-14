package main

// Client is the target notification entity
type Client struct {
	ID      string
	Channel chan Event
}
