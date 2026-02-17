package auth

import (
	"encoding/json"
	"net/http"

	"github.com/manuelgomezsw/loopi-api/internal/application/dto/request"
	dtoresponse "github.com/manuelgomezsw/loopi-api/internal/application/dto/response"
	"github.com/manuelgomezsw/loopi-api/internal/application/service"
	"github.com/manuelgomezsw/loopi-api/internal/interface/middleware"
	"github.com/manuelgomezsw/loopi-api/internal/interface/response"
	apperrors "github.com/manuelgomezsw/loopi-api/pkg/errors"
)

// Handler handles authentication endpoints (login, get me).
type Handler struct {
	authService *service.AuthService
}

// NewHandler creates a new auth handler.
func NewHandler(authService *service.AuthService) *Handler {
	return &Handler{authService: authService}
}

// Login handles POST /api/auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req request.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if err == apperrors.ErrInvalidCredentials {
			response.RespondError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		response.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := dtoresponse.LoginResponse{
		Token: result.Token,
		Employee: dtoresponse.EmployeeResponse{
			ID:       result.Employee.ID,
			Username: result.Employee.Username,
			Name:     result.Employee.Name,
			LastName: result.Employee.LastName,
			FullName: result.Employee.FullName(),
			Role:     string(result.Employee.Role),
		},
	}

	response.RespondJSON(w, http.StatusOK, resp)
}

// GetMe handles GET /api/employees/me.
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	employeeID, ok := middleware.GetEmployeeID(r.Context())
	if !ok {
		response.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	employee, err := h.authService.GetEmployeeByID(r.Context(), employeeID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			response.RespondError(w, http.StatusNotFound, "employee not found")
			return
		}
		response.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := dtoresponse.EmployeeResponse{
		ID:       employee.ID,
		Username: employee.Username,
		Name:     employee.Name,
		LastName: employee.LastName,
		FullName: employee.FullName(),
		Role:     string(employee.Role),
	}

	response.RespondJSON(w, http.StatusOK, resp)
}
