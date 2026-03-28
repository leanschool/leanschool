package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

// SchoolClassHandler handles CRUD for school classes.
// All routes require the school-management role.
type SchoolClassHandler struct {
	store storage.Storage
}

func NewSchoolClassHandler(store storage.Storage) *SchoolClassHandler {
	return &SchoolClassHandler{store: store}
}

func (h *SchoolClassHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /school-classes", h.Create)
	mux.HandleFunc("GET /school-classes", h.List)
	mux.HandleFunc("GET /school-classes/{id}", h.Get)
	mux.HandleFunc("PUT /school-classes/{id}", h.Update)
	mux.HandleFunc("DELETE /school-classes/{id}", h.Delete)
	mux.HandleFunc("GET /registration/school-classes", h.ListForRegistration)
}

func (h *SchoolClassHandler) requireWrite(w http.ResponseWriter, r *http.Request) bool {
	if !hasRole(ClaimsFromContext(r.Context()), "schoolclass_write_all") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}
	return true
}

func (h *SchoolClassHandler) requireRead(w http.ResponseWriter, r *http.Request) bool {
	if !hasRole(ClaimsFromContext(r.Context()), "schoolclass_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}
	return true
}

func (h *SchoolClassHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !h.requireWrite(w, r) {
		return
	}
	var sc model.SchoolClass
	if err := json.NewDecoder(r.Body).Decode(&sc); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreateSchoolClass(r.Context(), &sc); err != nil {
		http.Error(w, "create failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sc)
}

func (h *SchoolClassHandler) List(w http.ResponseWriter, r *http.Request) {
	if !h.requireRead(w, r) {
		return
	}
	classes, err := h.store.ListSchoolClasses(r.Context())
	if err != nil {
		http.Error(w, "list failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if classes == nil {
		classes = []*model.SchoolClass{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(classes)
}

// ListForRegistration returns id+name of all school classes for any authenticated user.
// Used by the registration form before the user has been assigned a role.
func (h *SchoolClassHandler) ListForRegistration(w http.ResponseWriter, r *http.Request) {
	if ClaimsFromContext(r.Context()) == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	classes, err := h.store.ListSchoolClasses(r.Context())
	if err != nil {
		http.Error(w, "list failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	type classRef struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	result := make([]classRef, 0, len(classes))
	for _, c := range classes {
		result = append(result, classRef{ID: c.ID, Name: c.Name})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *SchoolClassHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !h.requireRead(w, r) {
		return
	}
	id := r.PathValue("id")
	sc, err := h.store.GetSchoolClass(r.Context(), id)
	if err != nil {
		http.Error(w, "get failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if sc == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sc)
}

func (h *SchoolClassHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	id := r.PathValue("id")
	existing, err := h.store.GetSchoolClass(r.Context(), id)
	if err != nil || existing == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !hasWriteAccess(claims, "schoolclass", func() bool {
		for _, teacher := range existing.Teachers {
			if teacher.Sub != nil && *teacher.Sub == claims.Sub {
				return true
			}
		}
		return false
	}) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var sc model.SchoolClass
	if err := json.NewDecoder(r.Body).Decode(&sc); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	sc.ID = id
	if err := h.store.UpdateSchoolClass(r.Context(), &sc); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		if err.Error() == "school_classes not found" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "update failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sc)
}

func (h *SchoolClassHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	id := r.PathValue("id")
	existing, err := h.store.GetSchoolClass(r.Context(), id)
	if err != nil || existing == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !hasWriteAccess(claims, "schoolclass", func() bool {
		for _, teacher := range existing.Teachers {
			if teacher.Sub != nil && *teacher.Sub == claims.Sub {
				return true
			}
		}
		return false
	}) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.DeleteSchoolClass(r.Context(), id); err != nil {
		if err.Error() == "school_classes not found" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "delete failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
