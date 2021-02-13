package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	broker := NewBroker()
	broker.Run()

	api := NewNotificationAPI(broker)

	router := mux.NewRouter()
	router.HandleFunc("/notify", api.NotifyEventHandler).Methods("POST")
	router.HandleFunc("/notifications/{clientID}", api.GetNotificationsHandler)
	router.HandleFunc("/notifications/{clientID}/stream", api.NotificationStreamingHandler)

	server := &http.Server{
		Handler: router,
		Addr:    "127.0.0.1:8000",
	}

	log.Fatal("HTTP server error: ", server.ListenAndServe())
}
