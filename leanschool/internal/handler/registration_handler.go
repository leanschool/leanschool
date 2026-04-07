package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
	model "github.com/Joel-Haeberli/leanschool-model"
)

type RegistrationHandler struct {
	store storage.Storage
}

func NewRegistrationHandler(store storage.Storage) *RegistrationHandler {
	return &RegistrationHandler{store: store}
}

func (h *RegistrationHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /registration/start", h.StartRegistration)
	mux.HandleFunc("GET /registration/workflow", h.ListWorkflows)
	mux.HandleFunc("GET /registration/workflow/{id}", h.GetWorkflow)
	mux.HandleFunc("POST /registration/{id}/approve", h.Approve)
	mux.HandleFunc("POST /registration/{id}/reject", h.Reject)
	mux.HandleFunc("POST /registration/{id}/cancel", h.Cancel)
}

// Start new registration process
func (h *RegistrationHandler) StartRegistration(w http.ResponseWriter, r *http.Request) {
	// This endpoint is public - no auth required
	
	var req model.RegistrationRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if len(req.DesiredRoles) == 0 {
		http.Error(w, "desiredRoles is required", http.StatusBadRequest)
		return
	}
	if req.PersonData.Name == "" || req.PersonData.Prename == "" {
		http.Error(w, "personData.name and personData.prename are required", http.StatusBadRequest)
		return
	}
	
	// Validate roles
	roleMappings, err := h.store.ListRoleMappings(r.Context())
	if err != nil {
		http.Error(w, "role validation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	validRoles := make(map[string]bool)
	for _, mapping := range roleMappings {
		validRoles[mapping.KeycloakRole] = true
	}
	
	for _, role := range req.DesiredRoles {
		if !validRoles[role] {
			http.Error(w, "invalid role: "+role, http.StatusBadRequest)
			return
		}
	}
	
	// Create user registry entry with temporary user_sub for registration
	// Use email as temporary identifier to ensure uniqueness during registration
	tempUserSub := "registration:" + req.ContactEmail
	user := &model.UserRegistry{
		UserSub:             tempUserSub,
		Email:               req.ContactEmail,
		RegistrationStatus: model.RegistrationStatusPending,
		KeycloakRoles:      req.DesiredRoles, // Set initial roles from registration request
		LocalRoles:         req.DesiredRoles, // Set initial local roles from registration request
	}
	
	if err := h.store.CreateUserRegistry(r.Context(), user); err != nil {
		http.Error(w, "user creation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Create workflow entry
	workflow := &model.RegistrationWorkflow{
		UserID:           user.ID,
		RequestType:      "multi", // or specific type based on roles
		RequestData:      map[string]interface{}{
			"personData":   req.PersonData,
			"teacherData":  req.TeacherData,
			"studentData":  req.StudentData,
			"guardianData": req.GuardianData,
			"contactEmail": req.ContactEmail,
		},
		DesiredRoles:     req.DesiredRoles,
		CurrentRoles:     []string{}, // Start with empty current roles
		ApprovalStatus:   string(model.RegistrationStatusPending),
		RequiresManualApproval: true, // Default - can be configured per role
	}
	
	if err := h.store.CreateRegistrationWorkflow(r.Context(), workflow); err != nil {
		http.Error(w, "workflow creation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "pending",
		"message":       "Registration submitted for approval",
		"workflowId":    workflow.ID,
		"nextSteps":     "Wait for admin approval",
	})
}

// List registration workflows
func (h *RegistrationHandler) ListWorkflows(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "user_management_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	statusFilter := r.URL.Query().Get("status")
	
	workflows, err := h.store.ListRegistrationWorkflows(r.Context(), statusFilter)
	if err != nil {
		log.Printf("Error listing registration workflows: %v", err)
		http.Error(w, "list failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workflows)
}

// Get specific workflow
func (h *RegistrationHandler) GetWorkflow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	
	// Check if user is admin or owner
	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	workflow, err := h.store.GetRegistrationWorkflow(r.Context(), id)
	if err != nil {
		http.Error(w, "get failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if workflow == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	
	// Check ownership or admin role
	if !hasRole(claims, "user_management_read") {
		// Check if current user is the owner
		user, err := h.store.GetUserRegistry(r.Context(), workflow.UserID)
		if err != nil || user == nil || user.UserSub != claims.Sub {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workflow)
}

// Approve registration
func (h *RegistrationHandler) Approve(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "user_management_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	id := r.PathValue("id")
	
	var req struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get workflow
	workflow, err := h.store.GetRegistrationWorkflow(r.Context(), id)
	if err != nil {
		http.Error(w, "get failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if workflow == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	
	if workflow.ApprovalStatus != string(model.RegistrationStatusPending) {
		http.Error(w, "workflow is not pending approval", http.StatusBadRequest)
		return
	}
	
	// Update workflow
	workflow.ApprovalStatus = string(model.RegistrationStatusApproved)
	workflow.ApprovalBy = ClaimsFromContext(r.Context()).Sub
	workflow.ApprovalAt = ptr(time.Now())
	workflow.ApprovalNotes = req.Notes
	
	if err := h.store.UpdateRegistrationWorkflow(r.Context(), workflow); err != nil {
		http.Error(w, "approval failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// TODO: Trigger post-approval actions (create person, assign roles, etc.)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workflow)
}

// Reject registration
func (h *RegistrationHandler) Reject(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "user_management_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	id := r.PathValue("id")
	
	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get workflow
	workflow, err := h.store.GetRegistrationWorkflow(r.Context(), id)
	if err != nil {
		http.Error(w, "get failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if workflow == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	
	if workflow.ApprovalStatus != string(model.RegistrationStatusPending) {
		http.Error(w, "workflow is not pending approval", http.StatusBadRequest)
		return
	}
	
	// Update workflow
	workflow.ApprovalStatus = string(model.RegistrationStatusDenied)
	workflow.RejectionReason = req.Reason
	workflow.RejectionBy = ClaimsFromContext(r.Context()).Sub
	workflow.RejectionAt = ptr(time.Now())
	
	if err := h.store.UpdateRegistrationWorkflow(r.Context(), workflow); err != nil {
		http.Error(w, "rejection failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workflow)
}

// Cancel registration
func (h *RegistrationHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	
	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get workflow
	workflow, err := h.store.GetRegistrationWorkflow(r.Context(), id)
	if err != nil {
		http.Error(w, "get failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if workflow == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	
	// Check ownership or admin role
	isAdmin := hasRole(claims, "user_management_write")
	if !isAdmin {
		// Check if current user is the owner
		user, err := h.store.GetUserRegistry(r.Context(), workflow.UserID)
		if err != nil || user == nil || user.UserSub != claims.Sub {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}
	
	if workflow.ApprovalStatus != string(model.RegistrationStatusPending) {
		http.Error(w, "only pending workflows can be cancelled", http.StatusBadRequest)
		return
	}
	
	// Update workflow
	workflow.ApprovalStatus = string(model.RegistrationStatusCancelled)
	
	if err := h.store.UpdateRegistrationWorkflow(r.Context(), workflow); err != nil {
		http.Error(w, "cancellation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workflow)
}

// Helper function to create time pointer
func ptr(t time.Time) *time.Time {
	return &t
}