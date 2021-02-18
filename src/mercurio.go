package main

import "fmt"

type Mercurio struct {
	JWTAuth JWTAuthMiddleware
	Broker  *Broker
	API     NotificationAPI
}

func NewMercurio() (*Mercurio, error) {
	// Basic underlying setup
	//

	authPrivateKey, err := GetAuthPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get a private key for JWT Auth Middleware due to: %s", err.Error())
	}

	jwtAuth, err := NewJWTAuthMiddleware(authPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT Auth Middleware due to: %s", err.Error())
	}

	databaseFilePath, err := GetDatabaseConnectionString()
	if err != nil {
		return nil, fmt.Errorf("failed to get file path for SQLite database due to: %s", err.Error())
	}

	database := ConnectSqliteDatabase(databaseFilePath, true)
	repository, err := NewSQLNotificationRepository(database)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification repository on top of an SQLite database due to: %s", err.Error())
	}

	broker := NewBroker(repository)
	api := NewNotificationAPI(broker, repository)

	mercurio := &Mercurio{
		JWTAuth: jwtAuth,
		Broker:  broker,
		API:     api,
	}

	return mercurio, nil
}
