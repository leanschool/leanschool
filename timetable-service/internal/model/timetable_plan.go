package model

import "time"

// PlanStatus represents the lifecycle state of a timetable plan.
type PlanStatus string

const (
	PlanStatusDraft     PlanStatus = "draft"
	PlanStatusPlanning  PlanStatus = "planning"
	PlanStatusResolving PlanStatus = "resolving"
	PlanStatusAccepted  PlanStatus = "accepted"
	PlanStatusFinalized PlanStatus = "finalized"
)

// TimetablePlan is the top-level aggregate for one planning cycle.
type TimetablePlan struct {
	ID           string     `json:"id"`
	SchoolYearID string     `json:"schoolYearId"`
	Name         string     `json:"name"`
	Status       PlanStatus `json:"status"`
	CreatedBy    string     `json:"createdBy"`
	CreatedAt    *time.Time `json:"createdAt,omitempty"`
	UpdatedAt    *time.Time `json:"updatedAt,omitempty"`
	Version      int        `json:"version"`
}
