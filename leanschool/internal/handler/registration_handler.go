package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

type RegistrationHandler struct {
	store storage.Storage
	kc    *KeycloakAdminClient
}

func NewRegistrationHandler(store storage.Storage, kc *KeycloakAdminClient) *RegistrationHandler {
	return &RegistrationHandler{store: store, kc: kc}
}

func (h *RegistrationHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /registration/start", h.StartRegistration)
	mux.HandleFunc("GET /registration/workflow", h.ListWorkflows)
	mux.HandleFunc("GET /registration/workflow/{id}", h.GetWorkflow)
	mux.HandleFunc("POST /registration/{id}/approve", h.Approve)
	mux.HandleFunc("POST /registration/{id}/reject", h.Reject)
	mux.HandleFunc("POST /registration/{id}/cancel", h.Cancel)
	mux.HandleFunc("GET /registration/school-classes", h.ListSchoolClassesForRegistration)
}

// Start new registration process (requires authentication)
func (h *RegistrationHandler) StartRegistration(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check for existing registration by this user
	existing, err := h.store.GetUserRegistryBySub(r.Context(), claims.Sub)
	if err != nil {
		log.Printf("[registration] GetUserRegistryBySub failed: %v", err)
		http.Error(w, "registration check failed", http.StatusInternalServerError)
		return
	}
	if existing != nil {
		// User already has a registration — return current status
		wf, err := h.store.GetLatestWorkflowByUserID(r.Context(), existing.ID)
		if err != nil {
			log.Printf("[registration] GetLatestWorkflowByUserID failed: %v", err)
			http.Error(w, "registration check failed", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		status := "pending"
		if wf != nil {
			status = wf.ApprovalStatus
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     status,
			"message":    "Registration already exists",
			"workflowId": wf.ID,
		})
		return
	}

	var req model.RegistrationRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
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
		log.Printf("[registration] ListRoleMappings failed: %v", err)
		http.Error(w, "role validation failed", http.StatusInternalServerError)
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

	// Use the real Keycloak sub from JWT
	user := &model.UserRegistry{
		UserSub:            claims.Sub,
		Email:              req.ContactEmail,
		RegistrationStatus: model.RegistrationStatusPending,
		KeycloakRoles:      req.DesiredRoles,
		LocalRoles:         req.DesiredRoles,
	}

	if err := h.store.CreateUserRegistry(r.Context(), user); err != nil {
		log.Printf("[registration] CreateUserRegistry failed: %v", err)
		http.Error(w, "registration failed", http.StatusInternalServerError)
		return
	}

	// Create workflow entry
	workflow := &model.RegistrationWorkflow{
		UserID:      user.ID,
		RequestType: "multi",
		RequestData: map[string]interface{}{
			"personData":   req.PersonData,
			"teacherData":  req.TeacherData,
			"studentData":  req.StudentData,
			"guardianData": req.GuardianData,
			"contactEmail": req.ContactEmail,
		},
		DesiredRoles:           req.DesiredRoles,
		CurrentRoles:           []string{},
		ApprovalStatus:         string(model.RegistrationStatusPending),
		RequiresManualApproval: true,
	}

	if err := h.store.CreateRegistrationWorkflow(r.Context(), workflow); err != nil {
		log.Printf("[registration] CreateRegistrationWorkflow failed: %v", err)
		http.Error(w, "registration failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "pending",
		"message":    "Registration submitted for approval",
		"workflowId": workflow.ID,
		"nextSteps":  "Wait for admin approval",
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
		http.Error(w, "list failed", http.StatusInternalServerError)
		return
	}

	// Enrich workflows with email from user_registry for display
	type enrichedWorkflow struct {
		*model.RegistrationWorkflow
		Email string `json:"email"`
	}

	result := make([]enrichedWorkflow, 0, len(workflows))
	for _, wf := range workflows {
		ew := enrichedWorkflow{RegistrationWorkflow: wf}
		if wf.UserID != "" {
			user, err := h.store.GetUserRegistry(r.Context(), wf.UserID)
			if err == nil && user != nil {
				ew.Email = user.Email
			}
		}
		result = append(result, ew)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Get specific workflow
func (h *RegistrationHandler) GetWorkflow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	workflow, err := h.store.GetRegistrationWorkflow(r.Context(), id)
	if err != nil {
		http.Error(w, "get failed", http.StatusInternalServerError)
		return
	}
	if workflow == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Check ownership or admin role
	if !hasRole(claims, "user_management_read") {
		user, err := h.store.GetUserRegistry(r.Context(), workflow.UserID)
		if err != nil || user == nil || user.UserSub != claims.Sub {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workflow)
}

// Approve registration — creates domain entities and assigns Keycloak roles
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
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Get workflow
	workflow, err := h.store.GetRegistrationWorkflow(r.Context(), id)
	if err != nil {
		log.Printf("[registration] GetRegistrationWorkflow failed: %v", err)
		http.Error(w, "get failed", http.StatusInternalServerError)
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

	// Get the user registry entry
	user, err := h.store.GetUserRegistry(r.Context(), workflow.UserID)
	if err != nil || user == nil {
		log.Printf("[registration] GetUserRegistry failed for userId=%s: %v", workflow.UserID, err)
		http.Error(w, "user not found", http.StatusInternalServerError)
		return
	}

	// ── Post-approval pipeline ──────────────────────────────────────────
	// CreateTeacher/Student/Guardian each create their own Person row internally,
	// so we only call CreatePerson for roles that don't have a dedicated entity.

	name, prename := extractPersonNames(workflow.RequestData)
	sub := user.UserSub
	var personID string
	needsPlainPerson := true

	for _, role := range workflow.DesiredRoles {
		switch role {
		case "teacher":
			teacher := extractTeacherFromRequestData(workflow.RequestData)
			teacher.Name = name
			teacher.Prename = prename
			teacher.Sub = &sub
			if err := h.store.CreateTeacher(r.Context(), &teacher); err != nil {
				log.Printf("[registration] CreateTeacher failed: %v", err)
				http.Error(w, "approval failed: could not create teacher", http.StatusInternalServerError)
				return
			}
			workflow.PendingTeacherID = teacher.ID
			personID = teacher.ID
			needsPlainPerson = false

		case "student":
			student := model.Student{Name: name, Prename: prename, Sub: &sub}
			if err := h.store.CreateStudent(r.Context(), &student); err != nil {
				log.Printf("[registration] CreateStudent failed: %v", err)
				http.Error(w, "approval failed: could not create student", http.StatusInternalServerError)
				return
			}
			workflow.PendingStudentID = student.ID
			personID = student.ID
			needsPlainPerson = false

		case "guardian":
			guardian := model.Guardian{Name: name, Prename: prename, Sub: &sub}
			if err := h.store.CreateGuardian(r.Context(), &guardian); err != nil {
				log.Printf("[registration] CreateGuardian failed: %v", err)
				http.Error(w, "approval failed: could not create guardian", http.StatusInternalServerError)
				return
			}
			workflow.PendingGuardianID = guardian.ID
			personID = guardian.ID
			needsPlainPerson = false
		}
	}

	// For roles without dedicated entities (school-management, admin, etc.), create a plain Person
	if needsPlainPerson {
		person := &model.Person{Name: name, Prename: prename, Sub: &sub}
		if err := h.store.CreatePerson(r.Context(), person); err != nil {
			log.Printf("[registration] CreatePerson failed: %v", err)
			http.Error(w, "approval failed: could not create person", http.StatusInternalServerError)
			return
		}
		personID = person.ID
	}

	// 3. Assign Keycloak roles (best-effort: log on failure but continue)
	if err := h.kc.AssignRealmRoles(r.Context(), user.UserSub, workflow.DesiredRoles); err != nil {
		log.Printf("[registration] AssignRealmRoles failed for sub=%s: %v", user.UserSub, err)
	}

	// 4. Update user registry: link person, set status to approved
	user.PersonID = personID
	user.RegistrationStatus = model.RegistrationStatusApproved
	user.UpdatedBy = ClaimsFromContext(r.Context()).Sub
	if err := h.store.UpdateUserRegistry(r.Context(), user); err != nil {
		log.Printf("[registration] UpdateUserRegistry failed: %v", err)
		http.Error(w, "approval failed: could not update user", http.StatusInternalServerError)
		return
	}

	// 5. Update workflow status
	now := time.Now()
	workflow.ApprovalStatus = string(model.RegistrationStatusApproved)
	workflow.ApprovalBy = ClaimsFromContext(r.Context()).Sub
	workflow.ApprovalAt = &now
	workflow.ApprovalNotes = req.Notes
	workflow.CurrentRoles = workflow.DesiredRoles
	workflow.CompletedAt = &now

	if err := h.store.UpdateRegistrationWorkflow(r.Context(), workflow); err != nil {
		log.Printf("[registration] UpdateRegistrationWorkflow failed: %v", err)
		http.Error(w, "approval failed: could not update workflow", http.StatusInternalServerError)
		return
	}

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
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Get workflow
	workflow, err := h.store.GetRegistrationWorkflow(r.Context(), id)
	if err != nil {
		http.Error(w, "get failed", http.StatusInternalServerError)
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
	now := time.Now()
	workflow.ApprovalStatus = string(model.RegistrationStatusDenied)
	workflow.RejectionReason = req.Reason
	workflow.RejectionBy = ClaimsFromContext(r.Context()).Sub
	workflow.RejectionAt = &now

	if err := h.store.UpdateRegistrationWorkflow(r.Context(), workflow); err != nil {
		http.Error(w, "rejection failed", http.StatusInternalServerError)
		return
	}

	// Update user registry status
	user, err := h.store.GetUserRegistry(r.Context(), workflow.UserID)
	if err == nil && user != nil {
		user.RegistrationStatus = model.RegistrationStatusDenied
		if updateErr := h.store.UpdateUserRegistry(r.Context(), user); updateErr != nil {
			log.Printf("[registration] UpdateUserRegistry (deny) failed: %v", updateErr)
		}
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
		http.Error(w, "get failed", http.StatusInternalServerError)
		return
	}
	if workflow == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Check ownership or admin role
	isAdmin := hasRole(claims, "user_management_write")
	if !isAdmin {
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
		http.Error(w, "cancellation failed", http.StatusInternalServerError)
		return
	}

	// Update user registry status
	user, err := h.store.GetUserRegistry(r.Context(), workflow.UserID)
	if err == nil && user != nil {
		user.RegistrationStatus = model.RegistrationStatusCancelled
		if updateErr := h.store.UpdateUserRegistry(r.Context(), user); updateErr != nil {
			log.Printf("[registration] UpdateUserRegistry (cancel) failed: %v", updateErr)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workflow)
}

// ListSchoolClassesForRegistration returns school classes for the registration wizard
func (h *RegistrationHandler) ListSchoolClassesForRegistration(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	classes, err := h.store.ListSchoolClasses(r.Context())
	if err != nil {
		log.Printf("[registration] ListSchoolClasses failed: %v", err)
		http.Error(w, "failed to list classes", http.StatusInternalServerError)
		return
	}

	type classDTO struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	result := make([]classDTO, 0, len(classes))
	for _, c := range classes {
		result = append(result, classDTO{ID: c.ID, Name: c.Name})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ── Helper functions for extracting data from requestData JSONB ──────────────

func extractPersonNames(data map[string]interface{}) (name, prename string) {
	if pd, ok := data["personData"]; ok {
		if m, ok := pd.(map[string]interface{}); ok {
			if v, ok := m["name"].(string); ok {
				name = v
			}
			if v, ok := m["prename"].(string); ok {
				prename = v
			}
		}
	}
	return
}

func extractTeacherFromRequestData(data map[string]interface{}) model.Teacher {
	t := model.Teacher{
		AtSchoolSince: time.Now(),
	}
	if td, ok := data["teacherData"]; ok && td != nil {
		if m, ok := td.(map[string]interface{}); ok {
			if v, ok := m["iban"].(string); ok {
				t.Iban = &v
			}
		}
	}
	return t
}

// Helper function to create time pointer
func ptr(t time.Time) *time.Time {
	return &t
}
