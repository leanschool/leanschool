package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

type TeacherHandler struct{ store storage.Storage }

func NewTeacherHandler(store storage.Storage) *TeacherHandler {
	return &TeacherHandler{store: store}
}

func (h *TeacherHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /teachers", h.Create)
	mux.HandleFunc("GET /teachers", h.List)
	mux.HandleFunc("GET /teachers/{id}", h.Get)
	mux.HandleFunc("PUT /teachers/{id}", h.Update)
	mux.HandleFunc("DELETE /teachers/{id}", h.Delete)
}

func (h *TeacherHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "teacher_write_all") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var t model.Teacher
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreateTeacher(r.Context(), &t); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func (h *TeacherHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "teacher_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	teachers, err := h.store.ListTeachers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if teachers == nil {
		teachers = []*model.Teacher{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teachers)
}

func (h *TeacherHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "teacher_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	t, err := h.store.GetTeacher(r.Context(), r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if t == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func (h *TeacherHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	var t model.Teacher
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	t.ID = r.PathValue("id")
	existing, err := h.store.GetTeacher(r.Context(), t.ID)
	if err != nil || existing == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !hasWriteAccess(claims, "teacher", func() bool { return existing.Sub != nil && claims.Sub == *existing.Sub }) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.UpdateTeacher(r.Context(), &t); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func (h *TeacherHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	id := r.PathValue("id")
	existing, err := h.store.GetTeacher(r.Context(), id)
	if err != nil || existing == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !hasWriteAccess(claims, "teacher", func() bool { return existing.Sub != nil && claims.Sub == *existing.Sub }) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.DeleteTeacher(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
