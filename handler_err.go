package main

import "net/http"

func errorHandler(w http.ResponseWriter, r *http.Request) {
	errorResponse(w, http.StatusInternalServerError, "An error occurred")
}
