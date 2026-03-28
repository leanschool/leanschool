package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

// UserHandler handles user registration, status, and profile endpoints.
type UserHandler struct {
	store storage.Storage
	kc    *KeycloakAdminClient
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(store storage.Storage, kc *KeycloakAdminClient) *UserHandler {
	return &UserHandler{store: store, kc: kc}
}

// RegisterRoutes registers all /users/ routes on the given mux.
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /users/register", h.Register)
	mux.HandleFunc("GET /users/me", h.Me)
	mux.HandleFunc("GET /users/role-mappings", h.ListRoleMappings)
	mux.HandleFunc("GET /users/profile", h.GetProfile)
	mux.HandleFunc("PUT /users/profile", h.UpsertProfile)
	mux.HandleFunc("GET /users/registrations", h.ListRegistrations)
	mux.HandleFunc("POST /users/registrations/{id}/approve", h.Approve)
	mux.HandleFunc("POST /users/registrations/{id}/deny", h.Deny)
	mux.HandleFunc("GET /users/teachers", h.ListTeachers)
	mux.HandleFunc("GET /registration/who-is-who", h.WhoIsWho)
}

// hasRole returns true if the claims contain the given role.
func hasRole(claims *jwtClaims, role string) bool {
	if claims == nil {
		return false
	}
	for _, r := range claims.RealmAccess.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// ── GET /users/teachers ────────────────────────────────────────────────────────

func (h *UserHandler) ListTeachers(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if !hasRole(claims, "user-management") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	users, err := h.kc.ListUsersWithRole(r.Context(), "teacher")
	if err != nil {
		log.Printf("[user-handler] ListUsersWithRole failed: %v", err)
		http.Error(w, "failed to fetch teachers", http.StatusInternalServerError)
		return
	}

	type teacherDTO struct {
		Sub  string `json:"sub"`
		Name string `json:"name"`
	}
	result := make([]teacherDTO, 0, len(users))
	for _, u := range users {
		name := strings.TrimSpace(u.FirstName + " " + u.LastName)
		if name == "" {
			name = u.Username
		}
		result = append(result, teacherDTO{Sub: u.ID, Name: name})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ── GET /registration/who-is-who ──────────────────────────────────────────────

func (h *UserHandler) WhoIsWho(w http.ResponseWriter, r *http.Request) {
	if ClaimsFromContext(r.Context()) == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	roles, err := h.kc.ListCompositeRoles(r.Context())
	if err != nil {
		log.Printf("[user-handler] WhoIsWho ListCompositeRoles failed: %v", err)
		http.Error(w, "failed to fetch roles", http.StatusInternalServerError)
		return
	}

	type memberDTO struct {
		ID        string `json:"id"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Username  string `json:"username"`
	}
	type roleGroup struct {
		Name    string      `json:"name"`
		Members []memberDTO `json:"members"`
	}

	result := make([]roleGroup, 0, len(roles))
	for _, role := range roles {
		users, err := h.kc.ListUsersWithRole(r.Context(), role.Name)
		if err != nil {
			log.Printf("[user-handler] WhoIsWho ListUsersWithRole(%s) failed: %v", role.Name, err)
			users = nil
		}
		members := make([]memberDTO, 0, len(users))
		for _, u := range users {
			members = append(members, memberDTO{
				ID: u.ID, FirstName: u.FirstName, LastName: u.LastName, Username: u.Username,
			})
		}
		result = append(result, roleGroup{Name: role.Name, Members: members})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ── GET /users/role-mappings ───────────────────────────────────────────────────

func (h *UserHandler) ListRoleMappings(w http.ResponseWriter, r *http.Request) {
	if ClaimsFromContext(r.Context()) == nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	options, err := h.kc.ListCompositeRoles(r.Context())
	if err != nil {
		log.Printf("[user-handler] ListCompositeRoles failed: %v", err)
		http.Error(w, "failed to fetch role mappings", http.StatusInternalServerError)
		return
	}
	if options == nil {
		options = []RoleOption{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(options)
}

// ── GET /users/me ──────────────────────────────────────────────────────────────

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	reg, err := h.store.GetRegistrationRequestByUserSub(r.Context(), claims.Sub)
	if err != nil {
		http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	status := model.UserStatus{
		RegistrationStatus: model.RegistrationStatus("none"),
		ProfileComplete:    false,
		ProfileSkipped:     false,
	}

	if reg != nil {
		status.RegistrationStatus = reg.Status

		profile, err := h.store.GetUserProfile(r.Context(), claims.Sub)
		if err != nil {
			http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if profile != nil {
			status.ProfileComplete = profile.ProfileComplete
			status.ProfileSkipped = profile.ProfileSkipped
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// ── POST /users/register ───────────────────────────────────────────────────────

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// If a request already exists, return current status.
	existing, err := h.store.GetRegistrationRequestByUserSub(r.Context(), claims.Sub)
	if err != nil {
		http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if existing != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(existing)
		return
	}

	var body struct {
		DesiredRoles []string `json:"desiredRoles"`
		Email        string   `json:"email"`
		ClassIDs     []string `json:"classIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(body.DesiredRoles) == 0 {
		http.Error(w, "desiredRoles must not be empty", http.StatusBadRequest)
		return
	}

	req := &model.RegistrationRequest{
		UserSub:      claims.Sub,
		Email:        body.Email,
		DesiredRoles: body.DesiredRoles,
		ClassIDs:     body.ClassIDs,
		Status:       model.RegistrationStatusPending,
		CreatedAt:    time.Now(),
	}

	if err := h.store.CreateRegistrationRequest(r.Context(), req); err != nil {
		http.Error(w, "create failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req)
}

// ── GET /users/profile ─────────────────────────────────────────────────────────

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	profile, err := h.store.GetUserProfile(r.Context(), claims.Sub)
	if err != nil {
		http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if profile == nil {
		profile = &model.UserProfile{UserSub: claims.Sub}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// ── PUT /users/profile ─────────────────────────────────────────────────────────

func (h *UserHandler) UpsertProfile(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var body struct {
		IBAN           string   `json:"iban"`
		Address        string   `json:"address"`
		Phone          string   `json:"phone"`
		ClassIDs       []string `json:"classIds"`
		ProfileSkipped bool     `json:"profileSkipped"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	profileComplete := body.IBAN != "" || body.Address != "" || body.Phone != ""

	profile := &model.UserProfile{
		UserSub:         claims.Sub,
		IBAN:            body.IBAN,
		Address:         body.Address,
		Phone:           body.Phone,
		ClassIDs:        body.ClassIDs,
		ProfileComplete: profileComplete,
		ProfileSkipped:  body.ProfileSkipped,
	}

	if err := h.store.UpsertUserProfile(r.Context(), profile); err != nil {
		http.Error(w, "upsert failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// ── GET /users/registrations ───────────────────────────────────────────────────

func (h *UserHandler) ListRegistrations(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if !hasRole(claims, "user-management") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	reqs, err := h.store.ListRegistrationRequests(r.Context())
	if err != nil {
		http.Error(w, "list failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if reqs == nil {
		reqs = []*model.RegistrationRequest{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reqs)
}

// ── POST /users/registrations/{id}/approve ─────────────────────────────────────

func (h *UserHandler) Approve(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if !hasRole(claims, "user-management") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	id := r.PathValue("id")

	req, err := h.store.GetRegistrationRequestByID(r.Context(), id)
	if err != nil {
		http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if req == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err := h.store.UpdateRegistrationRequestStatus(r.Context(), id, model.RegistrationStatusApproved); err != nil {
		http.Error(w, "update failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Assign roles in Keycloak (best-effort: log on failure but still return 200).
	if err := h.kc.AssignRealmRoles(r.Context(), req.UserSub, req.DesiredRoles); err != nil {
		log.Printf("[user-handler] AssignRealmRoles failed for sub=%s: %v", req.UserSub, err)
	}

	welcomeURL := "http://localhost:3000?welcome=approved"
	log.Printf("[user-handler] Welcome link for user %s: %s", req.UserSub, welcomeURL)

	w.WriteHeader(http.StatusOK)
}

// ── POST /users/registrations/{id}/deny ───────────────────────────────────────

func (h *UserHandler) Deny(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if !hasRole(claims, "user-management") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	id := r.PathValue("id")

	// Verify it exists first.
	req, err := h.store.GetRegistrationRequestByID(r.Context(), id)
	if err != nil {
		http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if req == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err := h.store.UpdateRegistrationRequestStatus(r.Context(), id, model.RegistrationStatusDenied); err != nil {
		http.Error(w, "update failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

