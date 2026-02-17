package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/manuelgomezsw/loopi-api/internal/application/service"
	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/interface/response"
)

// EmployeeHandler handles admin employee endpoints (CRUD, reset password, list active).
type EmployeeHandler struct {
	employee *service.AdminEmployeeService
}

// NewEmployeeHandler creates a new admin employee handler.
func NewEmployeeHandler(employee *service.AdminEmployeeService) *EmployeeHandler {
	return &EmployeeHandler{employee: employee}
}

// ListEmployees returns a paginated list of employees with filters.
func (h *EmployeeHandler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	filter := service.EmployeeFilter{Page: 1, PageSize: 20}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			filter.Page = p
		}
	}
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			filter.PageSize = ps
		}
	}
	if role := r.URL.Query().Get("role"); role != "" {
		rol := entity.Role(role)
		filter.Role = &rol
	}
	if activeStr := r.URL.Query().Get("active"); activeStr != "" {
		active := activeStr == "true"
		filter.Active = &active
	}
	filter.Search = r.URL.Query().Get("search")

	result, err := h.employee.ListEmployees(r.Context(), filter)
	if err != nil {
		response.RespondError(w, http.StatusInternalServerError, "failed to list employees")
		return
	}

	response.RespondJSON(w, http.StatusOK, result)
}

// GetEmployee returns a single employee by ID.
func (h *EmployeeHandler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "employeeID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid employee ID")
		return
	}

	employee, err := h.employee.GetEmployee(r.Context(), uint16(id))
	if err != nil {
		response.RespondError(w, http.StatusNotFound, "employee not found")
		return
	}

	response.RespondJSON(w, http.StatusOK, employee)
}

// CreateEmployeeRequest is the request body for creating an employee.
type CreateEmployeeRequest struct {
	Username       string  `json:"username"`
	Password       string  `json:"password"`
	Name           string  `json:"name"`
	LastName       string  `json:"last_name"`
	DocumentType   *string `json:"document_type"`
	DocumentNumber *string `json:"document_number"`
	Phone          *string `json:"phone"`
	Email          *string `json:"email"`
	BirthDate      *string `json:"birth_date"`
	Role           string  `json:"role"`
}

// CreateEmployee creates a new employee.
func (h *EmployeeHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var req CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var birthDate *time.Time
	if req.BirthDate != nil && *req.BirthDate != "" {
		t, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			response.RespondError(w, http.StatusBadRequest, "invalid birth_date format, use YYYY-MM-DD")
			return
		}
		birthDate = &t
	}

	serviceReq := service.CreateEmployeeRequest{
		Username:       req.Username,
		Password:       req.Password,
		Name:           req.Name,
		LastName:       req.LastName,
		DocumentType:   req.DocumentType,
		DocumentNumber: req.DocumentNumber,
		Phone:          req.Phone,
		Email:          req.Email,
		BirthDate:      birthDate,
		Role:           entity.Role(req.Role),
	}

	employee, err := h.employee.CreateEmployee(r.Context(), serviceReq)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusCreated, employee)
}

// UpdateEmployeeRequest is the request body for updating an employee.
type UpdateEmployeeRequest struct {
	Username       string  `json:"username"`
	Name           string  `json:"name"`
	LastName       string  `json:"last_name"`
	DocumentType   *string `json:"document_type"`
	DocumentNumber *string `json:"document_number"`
	Phone          *string `json:"phone"`
	Email          *string `json:"email"`
	BirthDate      *string `json:"birth_date"`
	Role           string  `json:"role"`
	Active         bool    `json:"active"`
}

// UpdateEmployee updates an existing employee.
func (h *EmployeeHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "employeeID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid employee ID")
		return
	}

	var req UpdateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var birthDate *time.Time
	if req.BirthDate != nil && *req.BirthDate != "" {
		t, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			response.RespondError(w, http.StatusBadRequest, "invalid birth_date format, use YYYY-MM-DD")
			return
		}
		birthDate = &t
	}

	serviceReq := service.UpdateEmployeeRequest{
		Username:       req.Username,
		Name:           req.Name,
		LastName:       req.LastName,
		DocumentType:   req.DocumentType,
		DocumentNumber: req.DocumentNumber,
		Phone:          req.Phone,
		Email:          req.Email,
		BirthDate:      birthDate,
		Role:           entity.Role(req.Role),
		Active:         req.Active,
	}

	employee, err := h.employee.UpdateEmployee(r.Context(), uint16(id), serviceReq)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, employee)
}

// UpdateEmployeeStatusRequest is the request body for updating employee status.
type UpdateEmployeeStatusRequest struct {
	Active bool `json:"active"`
}

// UpdateEmployeeStatus updates the active status of an employee.
func (h *EmployeeHandler) UpdateEmployeeStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "employeeID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid employee ID")
		return
	}

	var req UpdateEmployeeStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.employee.UpdateEmployeeStatus(r.Context(), uint16(id), req.Active); err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// ResetEmployeePassword resets an employee's password to the default.
func (h *EmployeeHandler) ResetEmployeePassword(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "employeeID")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		response.RespondError(w, http.StatusBadRequest, "invalid employee ID")
		return
	}

	if err := h.employee.ResetEmployeePassword(r.Context(), uint16(id)); err != nil {
		response.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// ListAllActiveEmployees returns all active employees for dropdowns.
func (h *EmployeeHandler) ListAllActiveEmployees(w http.ResponseWriter, r *http.Request) {
	employees, err := h.employee.ListAllActiveEmployees(r.Context())
	if err != nil {
		response.RespondError(w, http.StatusInternalServerError, "failed to list active employees")
		return
	}

	response.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"employees": employees,
		"total":     len(employees),
	})
}
