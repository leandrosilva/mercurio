package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/cors"
	"github.com/urfave/negroni"
)

func main() {
	LoadEnvironmentVars()

	// Basic underlying setup
	//

	mercurio, err := NewMercurio()
	if err != nil {
		log.Fatal(err)
	}

	jwtAuth := mercurio.JWTAuth
	broker := mercurio.Broker
	api := mercurio.API

	// HTTP Server Setup & Boot
	//

	n := negroni.Classic()

	c := cors.New(GetCORSOptions())
	n.Use(c)

	r := MountRoutes(jwtAuth, api)
	n.UseHandler(r)

	s := &http.Server{
		Addr:    GetHTTPServerAddress(),
		Handler: n,
	}

	// Spawns server on a goroutine in order to not block the flow
	go func() {
		log.Println("Running notification service broker")
		broker.Run()

		log.Println("HTTP server listening on", s.Addr)
		err := s.ListenAndServe()
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
	s.Shutdown(ctx)

	log.Println("Bye bye")
	os.Exit(0)
}
