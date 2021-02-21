package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
)

// NewHTTPServer creates a new HTTP server to serves Broker's Notification API
func NewHTTPServer(jwtAuth JWTAuthMiddleware, api NotificationAPI) (*http.Server, error) {
	n := negroni.Classic()

	c := cors.New(GetCORSOptions())
	n.Use(c)

	r := mountRoutes(jwtAuth, api)
	n.UseHandler(r)

	s := &http.Server{
		Addr:    GetHTTPServerAddress(),
		Handler: n,
	}

	return s, nil
}

func mountRoutes(jwtAuth JWTAuthMiddleware, api NotificationAPI) *mux.Router {
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

	return r
}
