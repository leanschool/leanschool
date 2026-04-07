package model

import "time"

type RegistrationStatus string

const (
	RegistrationStatusPending  RegistrationStatus = "pending"
	RegistrationStatusApproved RegistrationStatus = "approved"
	RegistrationStatusDenied   RegistrationStatus = "denied"
	RegistrationStatusActive   RegistrationStatus = "active"
	RegistrationStatusLegacy   RegistrationStatus = "legacy"
	RegistrationStatusSuspended RegistrationStatus = "suspended"
	RegistrationStatusExpired  RegistrationStatus = "expired"
	RegistrationStatusCancelled RegistrationStatus = "cancelled"
)

type RegistrationRequest struct {
	ID           string             `json:"id"`
	UserSub      string             `json:"userSub"`
	Email        string             `json:"email"`
	DesiredRoles []string           `json:"desiredRoles"`
	ClassIDs     []string           `json:"classIds"`
	Status       RegistrationStatus `json:"status"`
	CreatedAt    time.Time          `json:"createdAt"`
}

type UserProfile struct {
	UserSub         string   `json:"userSub"`
	IBAN            string   `json:"iban"`
	Address         string   `json:"address"`
	Phone           string   `json:"phone"`
	ClassIDs        []string `json:"classIds"`
	ProfileComplete bool     `json:"profileComplete"`
	ProfileSkipped  bool     `json:"profileSkipped"`
}

type UserStatus struct {
	RegistrationStatus RegistrationStatus `json:"registrationStatus"` // "none"|"pending"|"approved"|"denied"
	RejectionReason    string             `json:"rejectionReason,omitempty"`
	ProfileComplete    bool               `json:"profileComplete"`
	ProfileSkipped     bool               `json:"profileSkipped"`
}

// UserRegistry represents a user in the unified user registry
type UserRegistry struct {
	ID              string    `json:"id"`
	UserSub         string    `json:"userSub"`
	PersonID        string    `json:"personId,omitempty"`
	Email           string    `json:"email,omitempty"`
	RegistrationStatus RegistrationStatus `json:"registrationStatus"`
	KeycloakRoles   []string  `json:"keycloakRoles,omitempty"`
	LocalRoles      []string  `json:"localRoles,omitempty"`
	LastRoleSync    *time.Time `json:"lastRoleSync,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	LastLoginAt     *time.Time `json:"lastLoginAt,omitempty"`
	CreatedBy        string    `json:"createdBy,omitempty"`
	UpdatedBy        string    `json:"updatedBy,omitempty"`
	LegacyProfileID string    `json:"legacyProfileId,omitempty"`
	LegacyRequestID  string    `json:"legacyRequestId,omitempty"`
}

// RegistrationWorkflow represents a user registration workflow
type RegistrationWorkflow struct {
	ID               string    `json:"id"`
	UserID           string    `json:"userId"`
	RequestType      string    `json:"requestType"`
	RequestData      map[string]interface{} `json:"requestData"`
	DesiredRoles     []string  `json:"desiredRoles"`
	CurrentRoles     []string  `json:"currentRoles,omitempty"`
	ApprovalStatus   string    `json:"approvalStatus"`
	ApprovalBy       string    `json:"approvalBy,omitempty"`
	ApprovalAt       *time.Time `json:"approvalAt,omitempty"`
	ApprovalNotes    string    `json:"approvalNotes,omitempty"`
	RejectionReason  string    `json:"rejectionReason,omitempty"`
	RejectionBy      string    `json:"rejectionBy,omitempty"`
	RejectionAt      *time.Time `json:"rejectionAt,omitempty"`
	RequiresManualApproval bool      `json:"requiresManualApproval"`
	AutoApprovalRules map[string]interface{} `json:"autoApprovalRules,omitempty"`
	PendingTeacherID string    `json:"pendingTeacherId,omitempty"`
	PendingStudentID string    `json:"pendingStudentId,omitempty"`
	PendingGuardianID string   `json:"pendingGuardianId,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	CompletedAt      *time.Time `json:"completedAt,omitempty"`
}

// RoleMapping represents a mapping between Keycloak roles and local roles
type RoleMapping struct {
	KeycloakRole     string `json:"keycloakRole"`
	LocalRole        string `json:"localRole"`
	Description      string `json:"description"`
	AutoAssign       bool   `json:"autoAssign"`
	AutoCreate       bool   `json:"autoCreate"`
	RequiresApproval bool   `json:"requiresApproval"`
	CanTeach         bool   `json:"canTeach"`
	CanManageClasses bool   `json:"canManageClasses"`
	CanManageUsers   bool   `json:"canManageUsers"`
	CanViewReports   bool   `json:"canViewReports"`
	ShowInRegistration bool `json:"showInRegistration"`
	RegistrationOrder int    `json:"registrationOrder"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// RegistrationRequestDTO represents a registration request from frontend
type RegistrationRequestDTO struct {
	DesiredRoles []string          `json:"desiredRoles"`
	PersonData   Person            `json:"personData"`
	TeacherData  *Teacher          `json:"teacherData,omitempty"`
	StudentData  *Student          `json:"studentData,omitempty"`
	GuardianData *Guardian         `json:"guardianData,omitempty"`
	ContactEmail string            `json:"contactEmail"`
}

// UserLinkRequest represents a request to link a user to a person
type UserLinkRequest struct {
	UserSub    string  `json:"userSub"`
	PersonData *Person `json:"personData,omitempty"`
}

// RoleUpdateRequest represents a request to update user roles
type RoleUpdateRequest struct {
	Add    []string `json:"add,omitempty"`
	Remove []string `json:"remove,omitempty"`
}
