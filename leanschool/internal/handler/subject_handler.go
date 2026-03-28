package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

type SubjectHandler struct{ store storage.Storage }

func NewSubjectHandler(store storage.Storage) *SubjectHandler {
	return &SubjectHandler{store: store}
}

func (h *SubjectHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /subjects", h.Create)
	mux.HandleFunc("GET /subjects", h.List)
	mux.HandleFunc("GET /subjects/{id}", h.Get)
	mux.HandleFunc("PUT /subjects/{id}", h.Update)
	mux.HandleFunc("DELETE /subjects/{id}", h.Delete)
}

func (h *SubjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "subject_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var s model.Subject
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreateSubject(r.Context(), &s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}

func (h *SubjectHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "subject_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	subjects, err := h.store.ListSubjects(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if subjects == nil {
		subjects = []*model.Subject{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subjects)
}

func (h *SubjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "subject_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	s, err := h.store.GetSubject(r.Context(), r.PathValue("id"))
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

func (h *SubjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "subject_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var s model.Subject
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	s.ID = r.PathValue("id")
	if err := h.store.UpdateSubject(r.Context(), &s); err != nil {
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

func (h *SubjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "subject_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.DeleteSubject(r.Context(), r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
