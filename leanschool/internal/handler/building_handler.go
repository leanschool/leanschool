package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

type BuildingHandler struct{ store storage.Storage }

func NewBuildingHandler(store storage.Storage) *BuildingHandler {
	return &BuildingHandler{store: store}
}

func (h *BuildingHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /buildings", h.Create)
	mux.HandleFunc("GET /buildings", h.List)
	mux.HandleFunc("GET /buildings/{id}", h.Get)
	mux.HandleFunc("PUT /buildings/{id}", h.Update)
	mux.HandleFunc("DELETE /buildings/{id}", h.Delete)
}

func (h *BuildingHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "building_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var b model.Building
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreateBuilding(r.Context(), &b); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(b)
}

func (h *BuildingHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "building_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	buildings, err := h.store.ListBuildings(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if buildings == nil {
		buildings = []*model.Building{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(buildings)
}

func (h *BuildingHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "building_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	b, err := h.store.GetBuilding(r.Context(), r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if b == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}

func (h *BuildingHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "building_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var b model.Building
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	b.ID = r.PathValue("id")
	if err := h.store.UpdateBuilding(r.Context(), &b); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}

func (h *BuildingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "building_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.DeleteBuilding(r.Context(), r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
