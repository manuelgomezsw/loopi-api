package http

import (
	"encoding/json"
	"net/http"
)

// JSON envia una respuesta con payload JSON y status code
func JSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

// OK (200) con datos
func OK(w http.ResponseWriter, payload interface{}) {
	JSON(w, http.StatusOK, payload)
}

// Created (201) con nuevo recurso
func Created(w http.ResponseWriter, payload interface{}) {
	JSON(w, http.StatusCreated, payload)
}

// BadRequest (400) con mensaje de error
func BadRequest(w http.ResponseWriter, message string) {
	JSON(w, http.StatusBadRequest, map[string]string{"error": message})
}

// Unauthorized (401)
func Unauthorized(w http.ResponseWriter, message string) {
	JSON(w, http.StatusUnauthorized, map[string]string{"error": message})
}

// Forbidden (403)
func Forbidden(w http.ResponseWriter, message string) {
	JSON(w, http.StatusForbidden, map[string]string{"error": message})
}

// NotFound (404)
func NotFound(w http.ResponseWriter, message string) {
	JSON(w, http.StatusNotFound, map[string]string{"error": message})
}

// ServerError (500)
func ServerError(w http.ResponseWriter, message string) {
	JSON(w, http.StatusInternalServerError, map[string]string{"error": message})
}

// NoContent (204)
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
