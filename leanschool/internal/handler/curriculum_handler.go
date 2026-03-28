package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

type CurriculumHandler struct{ store storage.Storage }

func NewCurriculumHandler(store storage.Storage) *CurriculumHandler {
	return &CurriculumHandler{store: store}
}

func (h *CurriculumHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /curricula", h.Create)
	mux.HandleFunc("GET /curricula", h.List)
	mux.HandleFunc("GET /curricula/{id}", h.Get)
	mux.HandleFunc("PUT /curricula/{id}", h.Update)
	mux.HandleFunc("DELETE /curricula/{id}", h.Delete)
}

func (h *CurriculumHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "curriculum_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var c model.Curriculum
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreateCurriculum(r.Context(), &c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

func (h *CurriculumHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "curriculum_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	curricula, err := h.store.ListCurricula(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if curricula == nil {
		curricula = []*model.Curriculum{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(curricula)
}

func (h *CurriculumHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "curriculum_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	c, err := h.store.GetCurriculum(r.Context(), r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if c == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func (h *CurriculumHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "curriculum_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var c model.Curriculum
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	c.ID = r.PathValue("id")
	if err := h.store.UpdateCurriculum(r.Context(), &c); err != nil {
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

func (h *CurriculumHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "curriculum_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.DeleteCurriculum(r.Context(), r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
