package http

import (
	"encoding/json"
	"log"
	"loopi-api/internal/delivery/http/rest"
	"loopi-api/internal/domain"
	"loopi-api/internal/usecase"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type EmployeeHandler struct {
	employeeUseCase usecase.EmployeeUseCase
}

func NewEmployeeHandler(employeeUseCase usecase.EmployeeUseCase) *EmployeeHandler {
	return &EmployeeHandler{employeeUseCase: employeeUseCase}
}

var employeeRequest struct {
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	DocumentType   string  `json:"document_type"`
	DocumentNumber string  `json:"document_number"`
	Birthdate      string  `json:"birthdate"`
	Phone          string  `json:"phone"`
	Email          string  `json:"email"`
	Position       string  `json:"position"`
	Salary         float64 `json:"salary"`
	StoreID        int     `json:"store_id"`
}

func (h *EmployeeHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	employees, err := h.employeeUseCase.GetAll()
	if err != nil {
		rest.HandleError(w, err)
		return
	}
	rest.OK(w, employees)
}

func (h *EmployeeHandler) GetByStore(w http.ResponseWriter, r *http.Request) {
	storeIDStr := chi.URLParam(r, "store_id")
	storeID, err := strconv.Atoi(storeIDStr)
	if err != nil {
		rest.BadRequest(w, "invalid store_id")
		return
	}

	users, err := h.employeeUseCase.GetByStore(storeID)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, users)
}

func (h *EmployeeHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		rest.HandleError(w, err)
	}

	employee, err := h.employeeUseCase.FindByID(id)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, employee)
}

func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := json.NewDecoder(r.Body).Decode(&employeeRequest); err != nil {
		log.Printf("error: %v", err)
		rest.BadRequest(w, "Invalid input")
		return
	}

	user := domain.User{
		FirstName:      employeeRequest.FirstName,
		LastName:       employeeRequest.LastName,
		DocumentType:   employeeRequest.DocumentType,
		DocumentNumber: employeeRequest.DocumentNumber,
		Birthdate:      employeeRequest.Birthdate,
		Phone:          employeeRequest.Phone,
		Email:          employeeRequest.Email,
		Position:       employeeRequest.Position,
		Salary:         employeeRequest.Salary,
	}

	if err := h.employeeUseCase.Create(user, employeeRequest.StoreID); err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.Created(w, map[string]string{"message": "Employee created"})
}

func (h *EmployeeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		rest.BadRequest(w, "Invalid JSON body")
		return
	}

	if err := h.employeeUseCase.Update(id, updates); err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.NoContent(w)
}

func (h *EmployeeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		rest.HandleError(w, err)
	}

	if err := h.employeeUseCase.Delete(id); err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.NoContent(w)
}

// GetActiveEmployees retrieves only active employees
func (h *EmployeeHandler) GetActiveEmployees(w http.ResponseWriter, r *http.Request) {
	employees, err := h.employeeUseCase.GetActiveEmployees()
	if err != nil {
		rest.HandleError(w, err)
		return
	}
	rest.OK(w, employees)
}

// GetByStoreAndActive retrieves active employees for a specific store
func (h *EmployeeHandler) GetByStoreAndActive(w http.ResponseWriter, r *http.Request) {
	storeIDStr := chi.URLParam(r, "store_id")
	storeID, err := strconv.Atoi(storeIDStr)
	if err != nil {
		rest.BadRequest(w, "invalid store_id")
		return
	}

	employees, err := h.employeeUseCase.GetByStoreAndActive(storeID)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, employees)
}
