package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const (
	BASE_URI = "/api/clients/123/notifications"
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

	databasePath, err := GetDatabaseConnectionString()
	if err != nil {
		panic(fmt.Sprintf("failed to delete test database '%s' due to: %s", databasePath, err.Error()))
	}
	os.Remove(databasePath)

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
}

// Helpers
//

func addHeaders(r *http.Request) {
	r.Header.Add("Authorization", "Bearer "+os.Getenv("TEST_TOKEN_USER_123"))
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
	r, err := http.NewRequest("GET", BASE_URI, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.Handler(jwtAuth.Secure(api.GetNotificationsHandler))

	handler.ServeHTTP(rr, r)

	assertStatusCode(t, rr, http.StatusUnauthorized)
}

func TestGetNotificationsHandler_ShouldGetEmptyList(t *testing.T) {
	r, err := http.NewRequest("GET", BASE_URI, nil)
	if err != nil {
		t.Fatal(err)
	}

	addHeaders(r)

	rr := httptest.NewRecorder()
	handler := http.Handler(jwtAuth.Secure(api.GetNotificationsHandler))

	handler.ServeHTTP(rr, r)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContent(t, rr, `{"notifications":[]}`)
}
