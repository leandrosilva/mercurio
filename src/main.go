package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
)

func main() {
	// Basic underlying setup
	//

	jwtAuth, err := NewJWTAuthMiddleware()
	if err != nil {
		panic(fmt.Sprintf("failed to create JWT Auth Middleware due to: %s", err.Error()))
	}

	database := ConnectSqliteDatabase("./db/mercurio.db", true)
	repository, err := NewSQLNotificationRepository(database)
	if err != nil {
		panic(fmt.Sprintf("failed to create notification repository on top of an SQLite database due to: %s", err.Error()))
	}

	broker := NewBroker(repository)
	api := NewNotificationAPI(broker, repository)

	// HTTP Routing
	//

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		content := map[string]string{"message": "Welcome to Mercurio"}
		respondWithSuccess(w, content)
	})

	eventsRouter := r.PathPrefix("/api/events").Subrouter()
	eventsRouter.Handle("/unicast", jwtAuth.Secure(api.UnicastEventHandler)).Methods("POST")
	eventsRouter.Handle("/broadcast", jwtAuth.Secure(api.BroadcastEventHandler)).Methods("POST")

	clientsRouter := r.PathPrefix("/api/clients/{clientID}").Subrouter()
	clientsRouter.Handle("/notifications/stream", jwtAuth.Secure(api.StreamNotificationsHandler))
	clientsRouter.Handle("/notifications", jwtAuth.Secure(api.GetNotificationsHandler))
	clientsRouter.Handle("/notifications/{notificationID:[0-9]+}", jwtAuth.Secure(api.GetNotificationHandler))
	clientsRouter.Handle("/notifications/{notificationID:[0-9]+}/read", jwtAuth.Secure(api.MarkNotificationReadHandler)).Methods("PUT")
	clientsRouter.Handle("/notifications/{notificationID:[0-9]+}/unread", jwtAuth.Secure(api.MarkNotificationUnreadHandler)).Methods("PUT")

	// CORS
	//

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"},
	})

	// HTTP Server Setup & Boot
	//

	n := negroni.Classic()
	n.Use(c)
	n.UseHandler(r)

	server := &http.Server{
		Addr:    "127.0.0.1:8000",
		Handler: n,
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
