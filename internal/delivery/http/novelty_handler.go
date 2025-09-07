package http

import (
	"encoding/json"
	"loopi-api/internal/delivery/http/rest"
	"loopi-api/internal/domain"
	"loopi-api/internal/usecase"
	"net/http"
	"strconv"
)

type NoveltyHandler struct {
	uc usecase.NoveltyUseCase
}

func NewNoveltyHandler(uc usecase.NoveltyUseCase) *NoveltyHandler {
	return &NoveltyHandler{uc}
}

func (h *NoveltyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.Novelty
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.BadRequest(w, "Invalid request body")
		return
	}
	if err := h.uc.Create(req); err != nil {
		rest.BadRequest(w, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	rest.OK(w, map[string]string{"message": "Novelty registered"})
}

func (h *NoveltyHandler) GetByEmployeeAndMonth(w http.ResponseWriter, r *http.Request) {
	employeeID, _ := strconv.Atoi(r.URL.Query().Get("employee"))
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))

	if employeeID == 0 || year == 0 || month == 0 {
		rest.BadRequest(w, "Missing parameters")
		return
	}

	novelties, err := h.uc.GetByEmployeeAndMonth(employeeID, year, month)
	if err != nil {
		rest.ServerError(w, err.Error())
		return
	}

	rest.OK(w, novelties)
}
