package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

// Mercurio is what you thing it is, or not
type Mercurio struct {
	// Node identification must be unique in a multi deployment setup
	NID string

	JWTAuth    JWTAuthMiddleware
	Broker     *Broker
	API        NotificationAPI
	HTTPServer *http.Server
}

// NewMercurio creates a new instance of Mercurio
func NewMercurio() (*Mercurio, error) {
	nid := GetNID()

	authPrivateKey, err := GetAuthPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get a private key for JWT Auth Middleware due to: %s", err)
	}

	jwtAuth, err := NewJWTAuthMiddleware(authPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT Auth Middleware due to: %s", err)
	}

	databaseFilePath, err := GetDatabaseConnectionString()
	if err != nil {
		return nil, fmt.Errorf("failed to get file path for SQLite database due to: %s", err)
	}

	database, err := ConnectSqliteDatabase(databaseFilePath, true)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite database due to: %s", err)
	}

	repository, err := NewSQLNotificationRepository(database)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification repository on top of an SQLite database due to: %s", err)
	}

	mqSettings, err := GetMQSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to get settings to connect to RabbitMQ server due to: %s", err)
	}

	broker, err := NewBroker(nid, repository, mqSettings)
	if err != nil {
		return nil, fmt.Errorf("failed to create Broker due to: %s", err)
	}

	api := NewNotificationAPI(broker, repository)

	httpServer, err := NewHTTPServer(jwtAuth, api)
	if err != nil {
		return nil, fmt.Errorf("failed to create the HTTP server due to: %s", err)
	}

	mercurio := &Mercurio{
		NID:        nid,
		JWTAuth:    jwtAuth,
		Broker:     broker,
		API:        api,
		HTTPServer: httpServer,
	}

	return mercurio, nil
}

// Start runs Broker, HTTP server, and everything else
func (m *Mercurio) Start() error {
	log.Println("Starting notification service broker")
	err := m.Broker.Run()
	if err != nil {
		return fmt.Errorf("failed running Broker due to: %s", err)
	}

	log.Println("HTTP server listening on", m.HTTPServer.Addr)
	err = m.HTTPServer.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed running HTTP server due to: %s", err)
	}

	return nil
}

// Stop the HTTP server, close MQ channel, and cleans everything before go
func (m *Mercurio) Stop(ctx context.Context) {
	log.Println("Stopping Broker")
	m.Broker.Stop()

	log.Println("Shutting down HTTP server")
	m.HTTPServer.Shutdown(ctx)
}
