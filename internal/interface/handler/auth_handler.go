package handler

import (
	"encoding/json"
	"net/http"

	"github.com/manuelgomezsw/loopi-api/internal/application/dto/request"
	"github.com/manuelgomezsw/loopi-api/internal/application/dto/response"
	"github.com/manuelgomezsw/loopi-api/internal/application/service"
	"github.com/manuelgomezsw/loopi-api/internal/interface/middleware"
	apperrors "github.com/manuelgomezsw/loopi-api/pkg/errors"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req request.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if err == apperrors.ErrInvalidCredentials {
			respondError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := response.LoginResponse{
		Token: result.Token,
		Employee: response.EmployeeResponse{
			ID:       result.Employee.ID,
			Username: result.Employee.Username,
			Name:     result.Employee.Name,
			LastName: result.Employee.LastName,
			FullName: result.Employee.FullName(),
			Role:     string(result.Employee.Role),
		},
	}

	respondJSON(w, http.StatusOK, resp)
}

// GetMe handles GET /api/employees/me.
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	employeeID, ok := middleware.GetEmployeeID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	employee, err := h.authService.GetEmployeeByID(r.Context(), employeeID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "employee not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := response.EmployeeResponse{
		ID:       employee.ID,
		Username: employee.Username,
		Name:     employee.Name,
		LastName: employee.LastName,
		FullName: employee.FullName(),
		Role:     string(employee.Role),
	}

	respondJSON(w, http.StatusOK, resp)
}
