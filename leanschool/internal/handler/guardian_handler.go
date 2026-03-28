package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

type GuardianHandler struct{ store storage.Storage }

func NewGuardianHandler(store storage.Storage) *GuardianHandler {
	return &GuardianHandler{store: store}
}

func (h *GuardianHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /guardians", h.Create)
	mux.HandleFunc("GET /guardians", h.List)
	mux.HandleFunc("GET /guardians/{id}", h.Get)
	mux.HandleFunc("PUT /guardians/{id}", h.Update)
	mux.HandleFunc("DELETE /guardians/{id}", h.Delete)
}

func (h *GuardianHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "guardian_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var g model.Guardian
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreateGuardian(r.Context(), &g); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(g)
}

func (h *GuardianHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "guardian_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	guardians, err := h.store.ListGuardians(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if guardians == nil {
		guardians = []*model.Guardian{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(guardians)
}

func (h *GuardianHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "guardian_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	g, err := h.store.GetGuardian(r.Context(), r.PathValue("id"))
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

func (h *GuardianHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "guardian_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var g model.Guardian
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	g.ID = r.PathValue("id")
	if err := h.store.UpdateGuardian(r.Context(), &g); err != nil {
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

func (h *GuardianHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "guardian_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.DeleteGuardian(r.Context(), r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
