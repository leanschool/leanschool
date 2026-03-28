package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

type PersonHandler struct{ store storage.Storage }

func NewPersonHandler(store storage.Storage) *PersonHandler {
	return &PersonHandler{store: store}
}

func (h *PersonHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /persons", h.Create)
	mux.HandleFunc("GET /persons", h.List)
	mux.HandleFunc("GET /persons/{id}", h.Get)
	mux.HandleFunc("PUT /persons/{id}", h.Update)
	mux.HandleFunc("DELETE /persons/{id}", h.Delete)
}

func (h *PersonHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "person_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var p model.Person
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreatePerson(r.Context(), &p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *PersonHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "person_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	persons, err := h.store.ListPersons(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if persons == nil {
		persons = []*model.Person{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(persons)
}

func (h *PersonHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "person_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	p, err := h.store.GetPerson(r.Context(), r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if p == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *PersonHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "person_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var p model.Person
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	p.ID = r.PathValue("id")
	if err := h.store.UpdatePerson(r.Context(), &p); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *PersonHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "person_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.DeletePerson(r.Context(), r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
