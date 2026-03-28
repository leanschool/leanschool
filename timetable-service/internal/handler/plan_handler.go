package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Joel-Haeberli/timetable-service/internal/model"
	"github.com/Joel-Haeberli/timetable-service/internal/storage"
)

// PlanHandler handles CRUD for TimetablePlan.
type PlanHandler struct{ store storage.Storage }

func NewPlanHandler(store storage.Storage) *PlanHandler {
	return &PlanHandler{store: store}
}

func (h *PlanHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /plans", h.Create)
	mux.HandleFunc("GET /plans", h.List)
	mux.HandleFunc("GET /plans/{id}", h.Get)
	mux.HandleFunc("PUT /plans/{id}", h.Update)
	mux.HandleFunc("DELETE /plans/{id}", h.Delete)
}

func (h *PlanHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var plan model.TimetablePlan
	if err := json.NewDecoder(r.Body).Decode(&plan); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	claims := ClaimsFromContext(r.Context())
	if claims != nil {
		plan.CreatedBy = claims.Sub
	}
	if err := h.store.CreatePlan(r.Context(), &plan); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(plan)
}

func (h *PlanHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	plans, err := h.store.ListPlans(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if plans == nil {
		plans = []*model.TimetablePlan{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plans)
}

func (h *PlanHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	plan, err := h.store.GetPlan(r.Context(), r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if plan == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

func (h *PlanHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var plan model.TimetablePlan
	if err := json.NewDecoder(r.Body).Decode(&plan); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	plan.ID = r.PathValue("id")
	if err := h.store.UpdatePlan(r.Context(), &plan); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

func (h *PlanHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	// Only allow deletion of draft plans
	plan, err := h.store.GetPlan(r.Context(), r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if plan == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if plan.Status != model.PlanStatusDraft {
		http.Error(w, "can only delete plans in draft status", http.StatusBadRequest)
		return
	}
	if err := h.store.DeletePlan(r.Context(), r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
