package handler

import (
	"encoding/json"
	"net/http"
)

// respondJSON writes a JSON response.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// respondError writes an error JSON response.
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// respondSuccess writes a success JSON response.
func respondSuccess(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"message": message})
}
