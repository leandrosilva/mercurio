package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	database := ConnectSqliteDatabase("./db/mercurio.db", true)
	repository, err := NewSQLNotificationRepository(database)
	if err != nil {
		panic("failed to create notification repository on top of an SQLite database")
	}

	broker := NewBroker(repository)
	api := NewNotificationAPI(broker, repository)

	router := mux.NewRouter()
	router.HandleFunc("/api/events/notify", api.NotifyEventHandler).Methods("POST")
	router.HandleFunc("/api/events/broadcast", api.BroadcastEventHandler).Methods("POST")
	router.HandleFunc("/api/clients/{clientID}/notifications/stream", api.StreamNotificationsHandler)
	router.HandleFunc("/api/clients/{clientID}/notifications", api.GetNotificationsHandler)
	router.HandleFunc("/api/clients/{clientID}/notifications/{notificationID}", api.GetNotificationHandler)
	router.HandleFunc("/api/clients/{clientID}/notifications/{notificationID}/read", api.MarkNotificationReadHandler).Methods("PUT")
	router.HandleFunc("/api/clients/{clientID}/notifications/{notificationID}/unread", api.MarkNotificationUnreadHandler).Methods("PUT")

	server := &http.Server{
		Addr:    "127.0.0.1:8000",
		Handler: router,
	}

	// Spawns server on a goroutine in order to not block the flow
	go func() {
		log.Println("Running notification service broker")
		broker.Run()

		log.Println("HTTP server listening on", server.Addr)
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("HTTP server error: ", err)
		}
	}()

	// Shutdown signal handle: SIGINT (Ctrl+C)
	// Note that SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt)

	<-shutdownSignal
	log.Println("Got shutdown signal")

	// When get a shutdown signal, sets a deadline to wait for
	waitDuration := time.Second * 10
	ctx, cancel := context.WithTimeout(context.Background(), waitDuration)
	defer cancel()

	log.Println("Shutting down...")
	server.Shutdown(ctx)

	log.Println("Bye bye")
	os.Exit(0)
}
