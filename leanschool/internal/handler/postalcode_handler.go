package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

type PostalCodeHandler struct{ store storage.Storage }

func NewPostalCodeHandler(store storage.Storage) *PostalCodeHandler {
	return &PostalCodeHandler{store: store}
}

func (h *PostalCodeHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /postal-codes", h.Create)
	mux.HandleFunc("GET /postal-codes", h.List)
	mux.HandleFunc("GET /postal-codes/{id}", h.Get)
	mux.HandleFunc("PUT /postal-codes/{id}", h.Update)
	mux.HandleFunc("DELETE /postal-codes/{id}", h.Delete)
}

func (h *PostalCodeHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "postalcode_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var pc model.PostalCode
	if err := json.NewDecoder(r.Body).Decode(&pc); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreatePostalCode(r.Context(), &pc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pc)
}

func (h *PostalCodeHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "postalcode_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	pcs, err := h.store.ListPostalCodes(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if pcs == nil {
		pcs = []*model.PostalCode{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pcs)
}

func (h *PostalCodeHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "postalcode_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	id := r.PathValue("id")
	pc, err := h.store.GetPostalCode(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if pc == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pc)
}

func (h *PostalCodeHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "postalcode_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	id := r.PathValue("id")
	var pc model.PostalCode
	if err := json.NewDecoder(r.Body).Decode(&pc); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	pc.ID = id
	if err := h.store.UpdatePostalCode(r.Context(), &pc); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pc)
}

func (h *PostalCodeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "postalcode_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	id := r.PathValue("id")
	if err := h.store.DeletePostalCode(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
