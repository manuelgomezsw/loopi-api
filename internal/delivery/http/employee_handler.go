package http

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"loopi-api/internal/delivery/http/rest"
	"loopi-api/internal/domain"
	"loopi-api/internal/usecase"
	"net/http"
	"strconv"
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
	Password       string  `json:"password"`
	Salary         float64 `json:"salary"`
	RoleID         int     `json:"role_id"`
	FranchiseID    int     `json:"franchise_id"`
}

func (h *EmployeeHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	employees, err := h.employeeUseCase.GetAll()
	if err != nil {
		rest.HandleError(w, err)
		return
	}
	rest.OK(w, employees)
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
		PasswordHash:   employeeRequest.Password,
	}

	if err := h.employeeUseCase.Create(user, employeeRequest.RoleID, employeeRequest.FranchiseID); err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.Created(w, map[string]string{"message": "Employee created"})
}

func (h *EmployeeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		rest.HandleError(w, err)
	}

	var input domain.User
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		rest.BadRequest(w, "Invalid JSON")
		return
	}

	input.ID = uint(id)

	if err := h.employeeUseCase.Update(&input); err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, input)
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
