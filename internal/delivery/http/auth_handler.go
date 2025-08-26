package http

import (
	"encoding/json"
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
