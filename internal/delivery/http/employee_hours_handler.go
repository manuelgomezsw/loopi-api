package http

import (
	"loopi-api/internal/delivery/http/rest"
	"loopi-api/internal/usecase"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type EmployeeHoursHandler struct {
	employeeHoursUseCase usecase.EmployeeHoursUseCase
}

func NewEmployeeHoursHandler(uc usecase.EmployeeHoursUseCase) *EmployeeHoursHandler {
	return &EmployeeHoursHandler{uc}
}

func (h *EmployeeHoursHandler) GetMonthlySummary(w http.ResponseWriter, r *http.Request) {
	employeeIDStr := chi.URLParam(r, "id")
	employeeID, err := strconv.Atoi(employeeIDStr)
	if err != nil || employeeID <= 0 {
		rest.BadRequest(w, "Invalid request body")
		return
	}
	year := time.Now().Year()
	month := int(time.Now().Month())

	if y := r.URL.Query().Get("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil {
			year = parsed
		}
	}
	if m := r.URL.Query().Get("month"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil {
			month = parsed
		}
	}

	summary, err := h.employeeHoursUseCase.GetMonthlySummary(employeeID, year, month)
	if err != nil {
		rest.HandleError(w, err)
		return
	}

	rest.OK(w, summary)
}
