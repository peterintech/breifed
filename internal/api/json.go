package api

import (
	"encoding/json"
	"log"
	"net/http"
)

func errorResponse(w http.ResponseWriter, status int, message string) {
	if status > 499 {
		http.Error(w, message, status)
		return
	}
	type errorResponse struct {
		Error string `json:"error"`
	}

	jsonResponse(w, status, errorResponse{Error: message})
}

func jsonResponse(w http.ResponseWriter, status int, data any) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal JSON: %v", err)
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(dataBytes)
}
