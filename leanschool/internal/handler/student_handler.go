package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

type StudentHandler struct{ store storage.Storage }

func NewStudentHandler(store storage.Storage) *StudentHandler {
	return &StudentHandler{store: store}
}

func (h *StudentHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /students", h.Create)
	mux.HandleFunc("GET /students", h.List)
	mux.HandleFunc("GET /students/{id}", h.Get)
	mux.HandleFunc("PUT /students/{id}", h.Update)
	mux.HandleFunc("DELETE /students/{id}", h.Delete)
}

func (h *StudentHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "student_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var s model.Student
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreateStudent(r.Context(), &s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}

func (h *StudentHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "student_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	students, err := h.store.ListStudents(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if students == nil {
		students = []*model.Student{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(students)
}

func (h *StudentHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "student_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	s, err := h.store.GetStudent(r.Context(), r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if s == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func (h *StudentHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "student_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var s model.Student
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	s.ID = r.PathValue("id")
	if err := h.store.UpdateStudent(r.Context(), &s); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func (h *StudentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "student_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.DeleteStudent(r.Context(), r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
