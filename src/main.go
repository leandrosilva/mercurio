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
	broker := NewBroker()

	api := NewNotificationAPI(broker)

	router := mux.NewRouter()
	router.HandleFunc("/notify", api.NotifyEventHandler).Methods("POST")
	router.HandleFunc("/broadcast", api.BrodcastEventHandler).Methods("POST")
	router.HandleFunc("/notifications/{clientID}", api.GetNotificationsHandler)
	router.HandleFunc("/notifications/{clientID}/stream", api.StreamNotificationsHandler)

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
