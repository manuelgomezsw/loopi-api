package http

import (
	"encoding/json"
	"loopi-api/internal/delivery/http/rest"
	"loopi-api/internal/domain"
	"loopi-api/internal/usecase"
	"net/http"
	"strconv"
)

type FranchiseHandler struct {
	franchiseUseCase usecase.FranchiseUseCase
}

func NewFranchiseHandler(franchiseUseCase usecase.FranchiseUseCase) *FranchiseHandler {
	return &FranchiseHandler{franchiseUseCase: franchiseUseCase}
}

func (h *FranchiseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.Franchise
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.BadRequest(w, "Invalid request body")
		return
	}
	if err := h.franchiseUseCase.Create(req); err != nil {
		rest.BadRequest(w, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	rest.OK(w, map[string]string{"message": "Franchise created"})
}

func (h *FranchiseHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	franchises, err := h.franchiseUseCase.GetAll()
	if err != nil {
		rest.ServerError(w, err.Error())
		return
	}

	rest.OK(w, franchises)
}

func (h *FranchiseHandler) GetById(w http.ResponseWriter, r *http.Request) {
	franchiseID, _ := strconv.Atoi(r.URL.Query().Get("employee"))
	if franchiseID == 0 {
		rest.BadRequest(w, "Missing parameters")
		return
	}

	franchise, err := h.franchiseUseCase.GetById(franchiseID)
	if err != nil {
		rest.ServerError(w, err.Error())
		return
	}

	rest.OK(w, franchise)
}
