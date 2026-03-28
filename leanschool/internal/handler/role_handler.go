package handler

import (
	"encoding/json"
	"net/http"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
	model "github.com/Joel-Haeberli/leanschool-model"
)

type RoleHandler struct {
	store storage.Storage
}

func NewRoleHandler(store storage.Storage) *RoleHandler {
	return &RoleHandler{store: store}
}

func (h *RoleHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /roles/mappings", h.ListMappings)
	mux.HandleFunc("GET /roles/mappings/{keycloakRole}", h.GetMapping)
	mux.HandleFunc("POST /roles/mappings", h.CreateMapping)
	mux.HandleFunc("PATCH /roles/mappings/{keycloakRole}", h.UpdateMapping)
	mux.HandleFunc("DELETE /roles/mappings/{keycloakRole}", h.DeleteMapping)
	mux.HandleFunc("POST /users/{id}/roles", h.AddRoles)
	mux.HandleFunc("DELETE /users/{id}/roles", h.RemoveRoles)
	mux.HandleFunc("POST /users/{id}/roles/sync", h.SyncRoles)
}

// List all role mappings
func (h *RoleHandler) ListMappings(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "user_management_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	mappings, err := h.store.ListRoleMappings(r.Context())
	if err != nil {
		http.Error(w, "list failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mappings)
}

// Get specific role mapping
func (h *RoleHandler) GetMapping(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "user_management_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	keycloakRole := r.PathValue("keycloakRole")
	
	mapping, err := h.store.GetRoleMapping(r.Context(), keycloakRole)
	if err != nil {
		http.Error(w, "get failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if mapping == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapping)
}

// Create role mapping
func (h *RoleHandler) CreateMapping(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "user_management_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	var mapping model.RoleMapping
	if err := json.NewDecoder(r.Body).Decode(&mapping); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	if mapping.KeycloakRole == "" || mapping.LocalRole == "" {
		http.Error(w, "keycloakRole and localRole are required", http.StatusBadRequest)
		return
	}
	
	if err := h.store.CreateRoleMapping(r.Context(), &mapping); err != nil {
		http.Error(w, "create failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mapping)
}

// Update role mapping
func (h *RoleHandler) UpdateMapping(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "user_management_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	keycloakRole := r.PathValue("keycloakRole")
	
	// Get existing mapping
	existing, err := h.store.GetRoleMapping(r.Context(), keycloakRole)
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
	
	// Apply updates
	if description, ok := updates["description"].(string); ok {
		existing.Description = description
	}
	if autoAssign, ok := updates["autoAssign"].(bool); ok {
		existing.AutoAssign = autoAssign
	}
	// Apply other fields similarly
	
	if err := h.store.UpdateRoleMapping(r.Context(), existing); err != nil {
		http.Error(w, "update failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existing)
}

// Delete role mapping
func (h *RoleHandler) DeleteMapping(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "user_management_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	keycloakRole := r.PathValue("keycloakRole")
	
	if err := h.store.DeleteRoleMapping(r.Context(), keycloakRole); err != nil {
		http.Error(w, "delete failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// Add roles to user
func (h *RoleHandler) AddRoles(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "user_management_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	id := r.PathValue("id")
	
	var req model.RoleUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	if len(req.Add) == 0 {
		http.Error(w, "no roles to add", http.StatusBadRequest)
		return
	}
	
	// Get user
	user, err := h.store.GetUserRegistry(r.Context(), id)
	if err != nil {
		http.Error(w, "get failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	
	// Validate roles exist in mappings
	mappings, err := h.store.ListRoleMappings(r.Context())
	if err != nil {
		http.Error(w, "role validation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	validRoles := make(map[string]bool)
	for _, mapping := range mappings {
		validRoles[mapping.LocalRole] = true
	}
	
	for _, role := range req.Add {
		if !validRoles[role] {
			http.Error(w, "invalid role: "+role, http.StatusBadRequest)
			return
		}
	}
	
	// Add roles (avoid duplicates)
	roleSet := make(map[string]bool)
	for _, role := range user.LocalRoles {
		roleSet[role] = true
	}
	
	for _, role := range req.Add {
		if !roleSet[role] {
			user.LocalRoles = append(user.LocalRoles, role)
			roleSet[role] = true
		}
	}
	
	if err := h.store.UpdateUserRegistry(r.Context(), user); err != nil {
		http.Error(w, "update failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Remove roles from user
func (h *RoleHandler) RemoveRoles(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "user_management_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	id := r.PathValue("id")
	
	var req model.RoleUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	if len(req.Remove) == 0 {
		http.Error(w, "no roles to remove", http.StatusBadRequest)
		return
	}
	
	// Get user
	user, err := h.store.GetUserRegistry(r.Context(), id)
	if err != nil {
		http.Error(w, "get failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	
	// Remove roles
	newRoles := []string{}
	roleSet := make(map[string]bool)
	for _, role := range req.Remove {
		roleSet[role] = true
	}
	
	for _, role := range user.LocalRoles {
		if !roleSet[role] {
			newRoles = append(newRoles, role)
		}
	}
	
	if len(newRoles) == len(user.LocalRoles) {
		// No roles were removed
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
		return
	}
	
	user.LocalRoles = newRoles
	
	if err := h.store.UpdateUserRegistry(r.Context(), user); err != nil {
		http.Error(w, "update failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Sync user roles with Keycloak
func (h *RoleHandler) SyncRoles(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	
	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Check if user is admin or self
	isAdmin := hasRole(claims, "user_management_write")
	if !isAdmin {
		// Check if it's the user's own account
		user, err := h.store.GetUserRegistry(r.Context(), id)
		if err != nil {
			http.Error(w, "get failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if user == nil || user.UserSub != claims.Sub {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}
	
	if err := h.store.SyncUserRoles(r.Context(), claims.Sub); err != nil {
		http.Error(w, "sync failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Get updated user
	user, err := h.store.GetUserRegistryBySub(r.Context(), claims.Sub)
	if err != nil {
		http.Error(w, "get failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"user":        user,
		"message":     "Roles synchronized successfully",
	})
}