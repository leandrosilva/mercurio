package main

import (
	"log"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

// JWTAuthMiddleware wrapper facility to an underlying JWTMiddleware
type JWTAuthMiddleware struct {
	handler *negroni.Negroni
}

// NewJWTAuthMiddleware creates a new JWTSecureMiddleware instance for our secret key
func NewJWTAuthMiddleware(privateKey []byte) (JWTAuthMiddleware, error) {
	middleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return privateKey, nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

	handler := negroni.New(negroni.HandlerFunc(middleware.HandlerWithNext))
	wrapper := JWTAuthMiddleware{handler: handler}

	return wrapper, nil
}

// Secure turns a otherwise public endpoint into a secure one
func (s *JWTAuthMiddleware) Secure(endpointHandler func(http.ResponseWriter, *http.Request)) *negroni.Negroni {
	return s.handler.With(
		negroni.HandlerFunc(checkAuthorizedUserIsValid),
		negroni.Wrap(http.HandlerFunc(endpointHandler)))
}

func checkAuthorizedUserIsValid(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if !isAuthorizationRequired(r) {
		next(w, r)
		return
	}

	vars := mux.Vars(r)

	// Is it a client route?
	clientID := vars["clientID"]
	if clientID != "" {
		// Does the token correspond to the expected client?
		if !isAuthorizedUserValid(r, clientID) {
			log.Printf("Blocking access: clientID %s", clientID)
			respondWithUnauthorized(w, "authorization token does not correspond to expected client")
			return
		}
	}

	// If its not a client route or client is valid, go ahead
	next(w, r)
}

func isAuthorizationRequired(r *http.Request) bool {
	return r.Method == "GET" || r.Method == "POST" || r.Method == "PUT"
}

func isAuthorizedUserValid(r *http.Request, expectedUserID string) bool {
	claims := decodeJWTClaims(r)
	userID := claims["user_id"]

	return userID == expectedUserID
}

func decodeJWTClaims(r *http.Request) jwt.MapClaims {
	user := r.Context().Value("user")
	if user == nil {
		log.Printf("failed to decode user from JWT token")
		return jwt.MapClaims{}
	}
	return user.(*jwt.Token).Claims.(jwt.MapClaims)
}
