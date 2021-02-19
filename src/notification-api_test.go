package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const (
	BASE_EVENTS_URL        = "/api/events"
	BASE_NOTIFICATIONS_URI = "/api/clients/123/notifications"
)

var (
	jwtAuth JWTAuthMiddleware
	broker  *Broker
	api     NotificationAPI
)

// Setup
//

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func setup() {
	LoadEnvironmentVars()

	deleteTestDatabase()

	// Basic underlying setup
	//

	mercurio, err := NewMercurio()
	if err != nil {
		panic(err)
	}

	jwtAuth = mercurio.JWTAuth
	api = mercurio.API
	broker = mercurio.Broker
	broker.Run()
}

func shutdown() {
	deleteTestDatabase()
}

// Helpers
//

func deleteTestDatabase() {
	databasePath, err := GetDatabaseConnectionString()
	if err != nil {
		panic(fmt.Sprintf("failed to delete test database '%s' due to: %s", databasePath, err.Error()))
	}
	os.Remove(databasePath)
}

func addUserAuthorization(r *http.Request, userID string) {
	r.Header.Add("Authorization", "Bearer "+os.Getenv("TEST_TOKEN_USER_"+userID))
}

func addPublisherAuthorization(r *http.Request, publisherID string) {
	r.Header.Add("Authorization", "Bearer "+os.Getenv("TEST_TOKEN_PUBLISHER_"+publisherID))
}

// Assertions
//

func assertStatusCode(t *testing.T, rr *httptest.ResponseRecorder, expected int) {
	if status := rr.Code; status != expected {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func assertBodyContent(t *testing.T, rr *httptest.ResponseRecorder, expected string) {
	body := strings.TrimSpace(rr.Body.String())
	if body != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			body, expected)
	}
}

func assertContent(t *testing.T, got interface{}, expected interface{}) {
	if got != expected {
		t.Errorf("unexpected value: got %v want %v",
			got, expected)
	}
}

// Test cases
//

func TestGetNotificationsHandler_WithoutAuthToken_ShouldBeUnauthorized(t *testing.T) {
	r, err := http.NewRequest("GET", BASE_NOTIFICATIONS_URI, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.Handler(jwtAuth.Secure(api.GetNotificationsHandler))

	handler.ServeHTTP(rr, r)

	assertStatusCode(t, rr, http.StatusUnauthorized)
}

func TestGetNotificationsHandler_WithoutNotificationsYet_ShouldGetEmptyList(t *testing.T) {
	r, err := http.NewRequest("GET", BASE_NOTIFICATIONS_URI, nil)
	if err != nil {
		t.Fatal(err)
	}

	addUserAuthorization(r, "123")

	rr := httptest.NewRecorder()
	handler := http.Handler(jwtAuth.Secure(api.GetNotificationsHandler))

	handler.ServeHTTP(rr, r)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContent(t, rr, `{"notifications":[]}`)
}

func TestUnicastNotificationHandler_WithoutAuthToken_ShouldBeUnauthorized(t *testing.T) {
	payload := `{"sourceID":"terminal","destinationID":"123","data":"some blah blah blah kind of thing"}`
	r, err := http.NewRequest("POST", BASE_EVENTS_URL+"/unicast", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.Handler(jwtAuth.Secure(api.UnicastEventHandler))

	handler.ServeHTTP(rr, r)

	assertStatusCode(t, rr, http.StatusUnauthorized)
}

func TestUnicastNotificationHandler(t *testing.T) {
	payload := `{"sourceID":"terminal","destinationID":"123","data":"some blah blah blah kind of thing"}`
	r, err := http.NewRequest("POST", BASE_EVENTS_URL+"/unicast", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}

	addPublisherAuthorization(r, "666")

	rr := httptest.NewRecorder()
	handler := http.Handler(jwtAuth.Secure(api.UnicastEventHandler))

	handler.ServeHTTP(rr, r)

	assertStatusCode(t, rr, http.StatusOK)

	body := strings.TrimSpace(rr.Body.String())
	object := make(map[string]interface{})
	err = json.Unmarshal([]byte(body), &object)
	if err != nil {
		t.Fatal(err)
	}
	assertContent(t, object["notificationID"], 1.0)
}

func TestGetNotificationHandler(t *testing.T) {
	r, err := http.NewRequest("GET", BASE_NOTIFICATIONS_URI+"/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	addUserAuthorization(r, "123")

	rr := httptest.NewRecorder()
	handler := http.Handler(jwtAuth.Secure(api.GetNotificationHandler))

	handler.ServeHTTP(rr, r)

	assertStatusCode(t, rr, http.StatusOK)

	body := strings.TrimSpace(rr.Body.String())
	object := make(map[string]interface{})
	err = json.Unmarshal([]byte(body), &object)
	if err != nil {
		t.Fatal(err)
	}
	assertContent(t, object["notificationID"], 1.0)
}
