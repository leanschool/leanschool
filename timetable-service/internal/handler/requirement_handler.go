package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Joel-Haeberli/timetable-service/internal/model"
	"github.com/Joel-Haeberli/timetable-service/internal/storage"
)

// RequirementHandler handles CRUD for GradeRequirement.
type RequirementHandler struct{ store storage.Storage }

func NewRequirementHandler(store storage.Storage) *RequirementHandler {
	return &RequirementHandler{store: store}
}

func (h *RequirementHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /plans/{planId}/requirements", h.Create)
	mux.HandleFunc("GET /plans/{planId}/requirements", h.List)
	mux.HandleFunc("PUT /plans/{planId}/requirements/{id}", h.Update)
	mux.HandleFunc("DELETE /plans/{planId}/requirements/{id}", h.Delete)
}

func (h *RequirementHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var req model.GradeRequirement
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.PlanID = r.PathValue("planId")
	if err := h.store.CreateRequirement(r.Context(), &req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req)
}

func (h *RequirementHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	reqs, err := h.store.ListRequirementsByPlan(r.Context(), r.PathValue("planId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if reqs == nil {
		reqs = []*model.GradeRequirement{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reqs)
}

func (h *RequirementHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var req model.GradeRequirement
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.ID = r.PathValue("id")
	req.PlanID = r.PathValue("planId")
	if err := h.store.UpdateRequirement(r.Context(), &req); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

func (h *RequirementHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.DeleteRequirement(r.Context(), r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
