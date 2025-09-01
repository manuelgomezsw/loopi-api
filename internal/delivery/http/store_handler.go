package http

import (
  "encoding/json"
  "github.com/go-chi/chi/v5"
  "loopi-api/internal/domain"
  "loopi-api/internal/repository"
  "net/http"
  "strconv"
)

type StoreHandler struct {
  repo repository.StoreRepository
}

func NewStoreHandler(repo repository.StoreRepository) *StoreHandler {
  return &StoreHandler{repo}
}

func (h *StoreHandler) GetAll(w http.ResponseWriter, r *http.Request) {
  stores, err := h.repo.GetAll()
  if err != nil {
    BadRequest(w, err.Error())
    return
  }

  OK(w, stores)
}

func (h *StoreHandler) GetByID(w http.ResponseWriter, r *http.Request) {
  id, _ := strconv.Atoi(chi.URLParam(r, "id"))
  store, err := h.repo.GetByID(id)
  if err != nil {
    NotFound(w, err.Error())
    return
  }

  OK(w, store)
}

func (h *StoreHandler) GetByFranchiseID(w http.ResponseWriter, r *http.Request) {
  franchiseIDStr := chi.URLParam(r, "franchiseID")
  franchiseID, err := strconv.Atoi(franchiseIDStr)
  if err != nil || franchiseID <= 0 {
    BadRequest(w, "Invalid franchise ID")
    return
  }

  stores, err := h.repo.GetByFranchiseID(franchiseID)
  if err != nil {
    ServerError(w, err.Error())
    return
  }

  OK(w, stores)
}

func (h *StoreHandler) Create(w http.ResponseWriter, r *http.Request) {
  var s domain.Store
  if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
    BadRequest(w, "Invalid JSON")
    return
  }
  if err := h.repo.Create(&s); err != nil {
    ServerError(w, err.Error())
    return
  }

  Created(w, s)
}

func (h *StoreHandler) Update(w http.ResponseWriter, r *http.Request) {
  id, _ := strconv.Atoi(chi.URLParam(r, "id"))
  var s domain.Store
  if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
    BadRequest(w, "Invalid JSON")
    return
  }
  s.ID = uint(id)
  if err := h.repo.Update(&s); err != nil {
    ServerError(w, err.Error())
    return
  }

  OK(w, s)
}

func (h *StoreHandler) Delete(w http.ResponseWriter, r *http.Request) {
  id, _ := strconv.Atoi(chi.URLParam(r, "id"))
  if err := h.repo.Delete(id); err != nil {
    ServerError(w, err.Error())
    return
  }

  NoContent(w)
}
