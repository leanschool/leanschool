package handler

import (
	"encoding/json"
	"net/http"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
	model "github.com/Joel-Haeberli/leanschool-model"
)

type UserRegistryHandler struct {
	store storage.Storage
}

func NewUserRegistryHandler(store storage.Storage) *UserRegistryHandler {
	return &UserRegistryHandler{store: store}
}

func (h *UserRegistryHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /users/registry", h.Create)
	mux.HandleFunc("GET /users/registry", h.List)
	mux.HandleFunc("GET /users/registry/{id}", h.Get)
	mux.HandleFunc("GET /users/registry/by-sub/{sub}", h.GetBySub)
	mux.HandleFunc("PATCH /users/registry/{id}", h.Update)
	mux.HandleFunc("POST /users/registry/{id}/link", h.Link)
}

func (h *UserRegistryHandler) requireUserManagementWrite(w http.ResponseWriter, r *http.Request) bool {
	if !hasRole(ClaimsFromContext(r.Context()), "user_management_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}
	return true
}

func (h *UserRegistryHandler) requireUserManagementRead(w http.ResponseWriter, r *http.Request) bool {
	if !hasRole(ClaimsFromContext(r.Context()), "user_management_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}
	return true
}

// Create user registry entry
func (h *UserRegistryHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !h.requireUserManagementWrite(w, r) {
		return
	}
	
	var user model.UserRegistry
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	if err := h.store.CreateUserRegistry(r.Context(), &user); err != nil {
		http.Error(w, "create failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// List all user registry entries
func (h *UserRegistryHandler) List(w http.ResponseWriter, r *http.Request) {
	if !h.requireUserManagementRead(w, r) {
		return
	}
	
	users, err := h.store.ListUserRegistries(r.Context())
	if err != nil {
		http.Error(w, "list failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// Get specific user registry entry
func (h *UserRegistryHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !h.requireUserManagementRead(w, r) {
		return
	}
	
	id := r.PathValue("id")
	user, err := h.store.GetUserRegistry(r.Context(), id)
	if err != nil {
		http.Error(w, "get failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Get user registry by Keycloak subject
func (h *UserRegistryHandler) GetBySub(w http.ResponseWriter, r *http.Request) {
	if !h.requireUserManagementRead(w, r) {
		return
	}
	
	sub := r.PathValue("sub")
	user, err := h.store.GetUserRegistryBySub(r.Context(), sub)
	if err != nil {
		http.Error(w, "get failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Update user registry entry
func (h *UserRegistryHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !h.requireUserManagementWrite(w, r) {
		return
	}
	
	id := r.PathValue("id")
	
	// Get existing user
	existing, err := h.store.GetUserRegistry(r.Context(), id)
	if err != nil {
		http.Error(w, "get failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if existing == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Apply updates to existing user
	if name, ok := updates["userSub"].(string); ok {
		existing.UserSub = name
	}
	if email, ok := updates["email"].(string); ok {
		existing.Email = email
	}
	if status, ok := updates["registrationStatus"].(string); ok {
		existing.RegistrationStatus = model.RegistrationStatus(status)
	}
	// Add other fields as needed
	
	if err := h.store.UpdateUserRegistry(r.Context(), existing); err != nil {
		http.Error(w, "update failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existing)
}

// Link user to person record
func (h *UserRegistryHandler) Link(w http.ResponseWriter, r *http.Request) {
	if !h.requireUserManagementWrite(w, r) {
		return
	}
	
	id := r.PathValue("id")
	
	var req struct {
		PersonID string `json:"personId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Get existing user
	user, err := h.store.GetUserRegistry(r.Context(), id)
	if err != nil {
		http.Error(w, "get failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	
	user.PersonID = req.PersonID
	
	if err := h.store.UpdateUserRegistry(r.Context(), user); err != nil {
		http.Error(w, "link failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}