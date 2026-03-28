package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

type LessonHandler struct{ store storage.Storage }

func NewLessonHandler(store storage.Storage) *LessonHandler {
	return &LessonHandler{store: store}
}

func (h *LessonHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /lessons", h.Create)
	mux.HandleFunc("GET /lessons", h.List)
	mux.HandleFunc("GET /lessons/{id}", h.Get)
	mux.HandleFunc("PUT /lessons/{id}", h.Update)
	mux.HandleFunc("DELETE /lessons/{id}", h.Delete)
}

func (h *LessonHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if !hasRole(claims, "lesson_write_own") && !hasRole(claims, "lesson_write_all") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var l model.Lesson
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreateLesson(r.Context(), &l); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(l)
}

func (h *LessonHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "lesson_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	lessons, err := h.store.ListLessons(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if lessons == nil {
		lessons = []*model.Lesson{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lessons)
}

func (h *LessonHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "lesson_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	l, err := h.store.GetLesson(r.Context(), r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if l == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(l)
}

func (h *LessonHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	id := r.PathValue("id")
	existing, err := h.store.GetLesson(r.Context(), id)
	if err != nil || existing == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !hasWriteAccess(claims, "lesson", func() bool {
		return existing.Teacher != nil && existing.Teacher.Sub != nil && *existing.Teacher.Sub == claims.Sub
	}) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var l model.Lesson
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	l.ID = id
	if err := h.store.UpdateLesson(r.Context(), &l); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(l)
}

func (h *LessonHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	id := r.PathValue("id")
	existing, err := h.store.GetLesson(r.Context(), id)
	if err != nil || existing == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !hasWriteAccess(claims, "lesson", func() bool {
		return existing.Teacher != nil && existing.Teacher.Sub != nil && *existing.Teacher.Sub == claims.Sub
	}) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.DeleteLesson(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
