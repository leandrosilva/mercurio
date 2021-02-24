package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	LoadEnvironmentVars()

	// Basic underlying setup
	//

	mercurio, err := NewMercurio()
	if err != nil {
		log.Fatal(err)
	}

	// Spawns server on a goroutine in order to not block the flow
	go func() {
		log.Printf("Mercurio %s is starting...", mercurio.NID)
		err := mercurio.Start()
		if err != nil {
			log.Fatalf("Mercurio %s failed due to: %s", mercurio.NID, err)
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
	mercurio.Stop(ctx)

	log.Println("Bye bye")
	os.Exit(0)
}
