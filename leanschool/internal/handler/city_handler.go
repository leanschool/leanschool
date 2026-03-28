package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

type CityHandler struct{ store storage.Storage }

func NewCityHandler(store storage.Storage) *CityHandler { return &CityHandler{store: store} }

func (h *CityHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /cities", h.Create)
	mux.HandleFunc("GET /cities", h.List)
	mux.HandleFunc("GET /cities/{id}", h.Get)
	mux.HandleFunc("PUT /cities/{id}", h.Update)
	mux.HandleFunc("DELETE /cities/{id}", h.Delete)
}

func (h *CityHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "city_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var c model.City
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreateCity(r.Context(), &c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

func (h *CityHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "city_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	cities, err := h.store.ListCities(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if cities == nil {
		cities = []*model.City{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cities)
}

func (h *CityHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "city_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	c, err := h.store.GetCity(r.Context(), r.PathValue("id"))
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

func (h *CityHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "city_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var c model.City
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	c.ID = r.PathValue("id")
	if err := h.store.UpdateCity(r.Context(), &c); err != nil {
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

func (h *CityHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "city_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.DeleteCity(r.Context(), r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
