package main

import (
	"encoding/json"
	"net/http"
)

// HTTP
//

func respondWithJSON(w http.ResponseWriter, content interface{}, httpStatus int) {
	response, err := json.Marshal(content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	w.Write(response)
}

func respondWithSuccess(w http.ResponseWriter, content interface{}) {
	respondWithJSON(w, content, http.StatusOK)
}

func respondWithError(w http.ResponseWriter, message string, httpStatus int) {
	content := map[string]string{"error": message}
	respondWithJSON(w, content, httpStatus)
}

func respondWithNotFound(w http.ResponseWriter, message string) {
	respondWithError(w, message, http.StatusNotFound)
}

func respondWithBadRequest(w http.ResponseWriter, message string) {
	respondWithError(w, message, http.StatusBadRequest)
}

func respondWithUnauthorized(w http.ResponseWriter, message string) {
	respondWithError(w, message, http.StatusUnauthorized)
}

func respondWithInternalServerError(w http.ResponseWriter, message string) {
	respondWithError(w, message, http.StatusInternalServerError)
}
