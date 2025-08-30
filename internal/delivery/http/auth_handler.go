package http

import (
	"encoding/json"
	"loopi-api/internal/middleware"
	"loopi-api/internal/usecase"
	nethttp "net/http"
)

type AuthHandler struct {
	authUseCase usecase.AuthUseCase
}

func NewAuthHandler(authUseCase usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUseCase: authUseCase}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ContextRequest struct {
	FranchiseID uint `json:"franchise_id"`
	StoreID     uint `json:"store_id"`
}

func (h *AuthHandler) Login(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}

	token, err := h.authUseCase.Login(req.Email, req.Password)
	if err != nil {
		Unauthorized(w, "Invalid credentials")
		return
	}

	OK(w, map[string]string{"token": token})
}

func (h *AuthHandler) SelectContext(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req ContextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.FranchiseID == 0 {
		BadRequest(w, "Missing or invalid context information")
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		Unauthorized(w, "Invalid token context")
		return
	}

	// Lógica de selección y verificación
	token, err := h.authUseCase.SelectContext(userID, req.FranchiseID, req.StoreID)
	if err != nil {
		Forbidden(w, err.Error())
		return
	}

	OK(w, map[string]string{"token": token})
}
