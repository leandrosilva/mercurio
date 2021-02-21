package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

const (
	baseEventsURL        = "/api/events"
	baseNotificationsURL = "/api/clients/{clientID}/notifications"
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
	ensureEnvironmentVars()

	LoadEnvironmentVars()

	deleteTestDatabase()

	// Gets the broker entity up & running
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

// General helpers
//

func ensureEnvironmentVars() {
	if os.Getenv("MERCURIO_ENV") == "" {
		os.Setenv("MERCURIO_ENV", "test")
	}

	if os.Getenv("MERCURIO_ENV_DIR") == "" {
		os.Setenv("MERCURIO_ENV_DIR", "../")
	}

	fmt.Printf(">> environment '%s' loaded from '%s'\n", os.Getenv("MERCURIO_ENV"), os.Getenv("MERCURIO_ENV_DIR"))
}

func deleteTestDatabase() {
	databasePath, err := GetDatabaseConnectionString()
	if err != nil {
		panic(fmt.Sprintf("failed to delete test database '%s' due to: %s", databasePath, err.Error()))
	}
	os.Remove(databasePath)
}

// HTTP req/res helpers
//

func addUserAuthorization(r *http.Request, userID string) {
	r.Header.Add("Authorization", "Bearer "+os.Getenv("TEST_TOKEN_USER_"+userID))
}

func addPublisherAuthorization(r *http.Request, publisherID string) {
	r.Header.Add("Authorization", "Bearer "+os.Getenv("TEST_TOKEN_PUBLISHER_"+publisherID))
}

func createUserRequest(t *testing.T, method string, url string, payload io.Reader) *http.Request {
	r, err := http.NewRequest(method, url, payload)
	if err != nil {
		t.Fatal(err)
	}

	addUserAuthorization(r, "123")

	return r
}

func createPublisherRequest(t *testing.T, method string, url string, payload io.Reader) *http.Request {
	r, err := http.NewRequest(method, url, payload)
	if err != nil {
		t.Fatal(err)
	}

	addPublisherAuthorization(r, "666")

	return r
}

func serveHTTPRequest(rt *mux.Router, r *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	rt.ServeHTTP(rr, r)

	return rr
}

func unmarshalBodyContent(t *testing.T, rr *httptest.ResponseRecorder) map[string]interface{} {
	body := strings.TrimSpace(rr.Body.String())
	object := map[string]interface{}{}

	err := json.Unmarshal([]byte(body), &object)
	if err != nil {
		fmt.Printf("|%v|", body)
		t.Fatal(err)
	}

	return object
}

// Assertions
//

func assertStatusCode(t *testing.T, rr *httptest.ResponseRecorder, expected int) {
	if status := rr.Code; status != expected {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, expected)
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
	r, err := http.NewRequest("GET", baseNotificationsURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.Handler(jwtAuth.Secure(api.GetNotificationsHandler))

	handler.ServeHTTP(rr, r)

	assertStatusCode(t, rr, http.StatusUnauthorized)
}

func TestUnicastNotificationHandler_WithoutAuthToken_ShouldBeUnauthorized(t *testing.T) {
	payload := `{"sourceID":"terminal","destinationID":"123","data":"some blah blah blah kind of thing"}`
	r, err := http.NewRequest("POST", baseEventsURL+"/unicast", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.Handler(jwtAuth.Secure(api.UnicastEventHandler))

	handler.ServeHTTP(rr, r)

	assertStatusCode(t, rr, http.StatusUnauthorized)
}

func TestHappyPathForUser123(t *testing.T) {
	rt := mux.NewRouter()
	rt.HandleFunc(baseEventsURL+"/unicast", jwtAuth.Secure(api.UnicastEventHandler).ServeHTTP)
	rt.HandleFunc(baseNotificationsURL, jwtAuth.Secure(api.GetNotificationsHandler).ServeHTTP)
	rt.HandleFunc(baseNotificationsURL+"/{notificationID}", jwtAuth.Secure(api.GetNotificationHandler).ServeHTTP)
	rt.HandleFunc(baseNotificationsURL+"/{notificationID}/read", jwtAuth.Secure(api.MarkNotificationReadHandler).ServeHTTP).Methods("PUT")
	rt.HandleFunc(baseNotificationsURL+"/{notificationID}/unread", jwtAuth.Secure(api.MarkNotificationUnreadHandler).ServeHTTP).Methods("PUT")

	baseNotificationsURL123 := strings.Replace(baseNotificationsURL, "{clientID}", "123", 1)

	// 1- Empty database, there is no notifications yet
	r := createUserRequest(t, "GET", baseNotificationsURL123, nil)
	rr := serveHTTPRequest(rt, r)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContent(t, rr, `{"clientID":"123","notifications":[]}`)

	// 2- Publishes one event to user
	payload := `{"sourceID":"test","destinationID":"123","data":"some blah blah blah kind of thing"}`
	r = createPublisherRequest(t, "POST", baseEventsURL+"/unicast", strings.NewReader(payload))
	rr = serveHTTPRequest(rt, r)

	assertStatusCode(t, rr, http.StatusOK)

	object := unmarshalBodyContent(t, rr)
	assertContent(t, object["notificationID"], 1.0)

	notificationID := object["notificationID"]
	eventID := object["eventID"]

	// 3- Gets the notification for the published event
	r = createUserRequest(t, "GET", fmt.Sprintf("%s/%v", baseNotificationsURL123, notificationID), nil)
	rr = serveHTTPRequest(rt, r)

	assertStatusCode(t, rr, http.StatusOK)

	object = unmarshalBodyContent(t, rr)
	assertContent(t, object["notificationID"], notificationID)
	assertContent(t, object["eventID"], eventID)

	// 4- Gets notifications, now, with the recently published one
	r = createUserRequest(t, "GET", baseNotificationsURL123, nil)
	rr = serveHTTPRequest(rt, r)

	assertStatusCode(t, rr, http.StatusOK)

	object = unmarshalBodyContent(t, rr)
	notifications := object["notifications"].([]interface{})
	assertContent(t, len(notifications), 1)

	notification1 := notifications[0].(map[string]interface{})
	assertContent(t, notification1["notificationID"], 1.0)

	// 5- Marks notification as read

	r = createUserRequest(t, "PUT", fmt.Sprintf("%s/%v/%s", baseNotificationsURL123, notificationID, "read"), nil)
	rr = serveHTTPRequest(rt, r)

	assertStatusCode(t, rr, http.StatusOK)

	object = unmarshalBodyContent(t, rr)
	assertBodyContent(t, rr, `{"status":"read"}`)

	// 5- Marks notification back as unread

	r = createUserRequest(t, "PUT", fmt.Sprintf("%s/%v/%s", baseNotificationsURL123, notificationID, "unread"), nil)
	rr = serveHTTPRequest(rt, r)

	assertStatusCode(t, rr, http.StatusOK)

	object = unmarshalBodyContent(t, rr)
	assertBodyContent(t, rr, `{"status":"unread"}`)
}
