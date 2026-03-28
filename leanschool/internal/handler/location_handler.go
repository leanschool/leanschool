package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

type LocationHandler struct{ store storage.Storage }

func NewLocationHandler(store storage.Storage) *LocationHandler {
	return &LocationHandler{store: store}
}

func (h *LocationHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /locations", h.Create)
	mux.HandleFunc("GET /locations", h.List)
	mux.HandleFunc("GET /locations/{id}", h.Get)
	mux.HandleFunc("PUT /locations/{id}", h.Update)
	mux.HandleFunc("DELETE /locations/{id}", h.Delete)
}

func (h *LocationHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "location_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var loc model.Location
	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreateLocation(r.Context(), &loc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(loc)
}

func (h *LocationHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "location_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	locs, err := h.store.ListLocations(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if locs == nil {
		locs = []*model.Location{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(locs)
}

func (h *LocationHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "location_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	loc, err := h.store.GetLocation(r.Context(), r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if loc == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loc)
}

func (h *LocationHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "location_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var loc model.Location
	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	loc.ID = r.PathValue("id")
	if err := h.store.UpdateLocation(r.Context(), &loc); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loc)
}

func (h *LocationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "location_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.DeleteLocation(r.Context(), r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
