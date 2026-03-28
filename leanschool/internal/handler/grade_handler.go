package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

type GradeHandler struct{ store storage.Storage }

func NewGradeHandler(store storage.Storage) *GradeHandler { return &GradeHandler{store: store} }

func (h *GradeHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /grades", h.Create)
	mux.HandleFunc("GET /grades", h.List)
	mux.HandleFunc("GET /grades/{id}", h.Get)
	mux.HandleFunc("PUT /grades/{id}", h.Update)
	mux.HandleFunc("DELETE /grades/{id}", h.Delete)
}

func (h *GradeHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if !hasRole(claims, "grade_write_own") && !hasRole(claims, "grade_write_all") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var g model.Grade
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreateGrade(r.Context(), &g); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(g)
}

func (h *GradeHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "grade_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	grades, err := h.store.ListGrades(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if grades == nil {
		grades = []*model.Grade{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(grades)
}

func (h *GradeHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "grade_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	g, err := h.store.GetGrade(r.Context(), r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if g == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(g)
}

func (h *GradeHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	id := r.PathValue("id")
	isTeacher, err := h.store.IsTeacherOfGrade(r.Context(), id, claims.Sub)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !hasWriteAccess(claims, "grade", func() bool { return isTeacher }) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var g model.Grade
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	g.ID = id
	if err := h.store.UpdateGrade(r.Context(), &g); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(g)
}

func (h *GradeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	id := r.PathValue("id")
	isTeacher, err := h.store.IsTeacherOfGrade(r.Context(), id, claims.Sub)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !hasWriteAccess(claims, "grade", func() bool { return isTeacher }) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.DeleteGrade(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
