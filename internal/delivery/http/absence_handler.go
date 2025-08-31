package http

import (
	"encoding/json"
	"loopi-api/internal/domain"
	"loopi-api/internal/usecase"
	"net/http"
	"strconv"
)

type AbsenceHandler struct {
	uc usecase.AbsenceUseCase
}

func NewAbsenceHandler(uc usecase.AbsenceUseCase) *AbsenceHandler {
	return &AbsenceHandler{uc}
}

func (h *AbsenceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.Absence
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}
	if err := h.uc.Create(req); err != nil {
		BadRequest(w, err.Error())
		return
	}

	OK(w, map[string]string{"message": "Absence registered"})
}

func (h *AbsenceHandler) GetByEmployeeAndMonth(w http.ResponseWriter, r *http.Request) {
	employeeID, _ := strconv.Atoi(r.URL.Query().Get("employee"))
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))

	if employeeID == 0 || year == 0 || month == 0 {
		BadRequest(w, "Missing parameters")
		return
	}

	absences, err := h.uc.GetByEmployeeAndMonth(employeeID, year, month)
	if err != nil {
		ServerError(w, err.Error())
		return
	}

	OK(w, absences)
}
