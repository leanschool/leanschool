package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

type SchoolYearHandler struct{ store storage.Storage }

func NewSchoolYearHandler(store storage.Storage) *SchoolYearHandler {
	return &SchoolYearHandler{store: store}
}

func (h *SchoolYearHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /school-years", h.Create)
	mux.HandleFunc("GET /school-years", h.List)
	mux.HandleFunc("GET /school-years/{id}", h.Get)
	mux.HandleFunc("PUT /school-years/{id}", h.Update)
	mux.HandleFunc("DELETE /school-years/{id}", h.Delete)
}

func (h *SchoolYearHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "schoolyear_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var sy model.SchoolYear
	if err := json.NewDecoder(r.Body).Decode(&sy); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreateSchoolYear(r.Context(), &sy); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sy)
}

func (h *SchoolYearHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "schoolyear_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	years, err := h.store.ListSchoolYears(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if years == nil {
		years = []*model.SchoolYear{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(years)
}

func (h *SchoolYearHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "schoolyear_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	sy, err := h.store.GetSchoolYear(r.Context(), r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if sy == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sy)
}

func (h *SchoolYearHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "schoolyear_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var sy model.SchoolYear
	if err := json.NewDecoder(r.Body).Decode(&sy); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	sy.ID = r.PathValue("id")
	if err := h.store.UpdateSchoolYear(r.Context(), &sy); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sy)
}

func (h *SchoolYearHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "schoolyear_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.DeleteSchoolYear(r.Context(), r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
