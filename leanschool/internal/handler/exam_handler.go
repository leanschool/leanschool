package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

type ExamHandler struct{ store storage.Storage }

func NewExamHandler(store storage.Storage) *ExamHandler { return &ExamHandler{store: store} }

func (h *ExamHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /exams", h.Create)
	mux.HandleFunc("GET /exams", h.List)
	mux.HandleFunc("GET /exams/{id}", h.Get)
	mux.HandleFunc("PUT /exams/{id}", h.Update)
	mux.HandleFunc("DELETE /exams/{id}", h.Delete)
}

func (h *ExamHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if !hasRole(claims, "exam_write_own") && !hasRole(claims, "exam_write_all") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var e model.Exam
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreateExam(r.Context(), &e); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(e)
}

func (h *ExamHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "exam_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	exams, err := h.store.ListExams(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exams == nil {
		exams = []*model.Exam{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exams)
}

func (h *ExamHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "exam_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	e, err := h.store.GetExam(r.Context(), r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if e == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

func (h *ExamHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	id := r.PathValue("id")
	isTeacher, err := h.store.IsTeacherOfExam(r.Context(), id, claims.Sub)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !hasWriteAccess(claims, "exam", func() bool { return isTeacher }) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var e model.Exam
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	e.ID = id
	if err := h.store.UpdateExam(r.Context(), &e); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

func (h *ExamHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	id := r.PathValue("id")
	isTeacher, err := h.store.IsTeacherOfExam(r.Context(), id, claims.Sub)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !hasWriteAccess(claims, "exam", func() bool { return isTeacher }) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.DeleteExam(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
