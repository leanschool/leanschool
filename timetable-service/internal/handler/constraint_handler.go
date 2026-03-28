package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Joel-Haeberli/timetable-service/internal/model"
	"github.com/Joel-Haeberli/timetable-service/internal/storage"
)

// ConstraintHandler handles CRUD for ClassConstraint.
type ConstraintHandler struct{ store storage.Storage }

func NewConstraintHandler(store storage.Storage) *ConstraintHandler {
	return &ConstraintHandler{store: store}
}

func (h *ConstraintHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /plans/{planId}/constraints", h.Create)
	mux.HandleFunc("GET /plans/{planId}/constraints", h.List)
	mux.HandleFunc("PUT /plans/{planId}/constraints/{id}", h.Update)
	mux.HandleFunc("DELETE /plans/{planId}/constraints/{id}", h.Delete)
}

func (h *ConstraintHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var c model.ClassConstraint
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	c.PlanID = r.PathValue("planId")
	if err := h.store.CreateConstraint(r.Context(), &c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

func (h *ConstraintHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	constraints, err := h.store.ListConstraintsByPlan(r.Context(), r.PathValue("planId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if constraints == nil {
		constraints = []*model.ClassConstraint{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(constraints)
}

func (h *ConstraintHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var c model.ClassConstraint
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	c.ID = r.PathValue("id")
	c.PlanID = r.PathValue("planId")
	if err := h.store.UpdateConstraint(r.Context(), &c); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func (h *ConstraintHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.DeleteConstraint(r.Context(), r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
