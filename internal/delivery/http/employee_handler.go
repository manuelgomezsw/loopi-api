package http

import (
	"encoding/json"
	"log"
	"loopi-api/internal/domain"
	"loopi-api/internal/usecase"
	"net/http"
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

func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := json.NewDecoder(r.Body).Decode(&employeeRequest); err != nil {
		log.Printf("❌ error: %v", err)
		BadRequest(w, "Invalid input")
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

	if err := h.employeeUseCase.CreateEmployee(user, employeeRequest.RoleID, employeeRequest.FranchiseID); err != nil {
		ServerError(w, err.Error())
		return
	}

	Created(w, map[string]string{"message": "Employee created"})
}
