package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func MountRoutes(jwtAuth JWTAuthMiddleware, api NotificationAPI) *mux.Router {
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
