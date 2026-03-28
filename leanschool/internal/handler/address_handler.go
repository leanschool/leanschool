package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

type AddressHandler struct{ store storage.Storage }

func NewAddressHandler(store storage.Storage) *AddressHandler {
	return &AddressHandler{store: store}
}

func (h *AddressHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /addresses", h.Create)
	mux.HandleFunc("GET /addresses", h.List)
	mux.HandleFunc("GET /addresses/{id}", h.Get)
	mux.HandleFunc("PUT /addresses/{id}", h.Update)
	mux.HandleFunc("DELETE /addresses/{id}", h.Delete)
}

func (h *AddressHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "address_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var a model.Address
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreateAddress(r.Context(), &a); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(a)
}

func (h *AddressHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "address_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	addrs, err := h.store.ListAddresses(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if addrs == nil {
		addrs = []*model.Address{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(addrs)
}

func (h *AddressHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "address_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	a, err := h.store.GetAddress(r.Context(), r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if a == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a)
}

func (h *AddressHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "address_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var a model.Address
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	a.ID = r.PathValue("id")
	if err := h.store.UpdateAddress(r.Context(), &a); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a)
}

func (h *AddressHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "address_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.DeleteAddress(r.Context(), r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
