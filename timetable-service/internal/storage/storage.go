package storage

import (
	"context"

	"github.com/Joel-Haeberli/timetable-service/internal/model"
)

// Storage defines CRUD operations for all timetable domain entities.
type Storage interface {
	// timetable plans
	CreatePlan(ctx context.Context, p *model.TimetablePlan) error
	GetPlan(ctx context.Context, id string) (*model.TimetablePlan, error)
	ListPlans(ctx context.Context) ([]*model.TimetablePlan, error)
	UpdatePlan(ctx context.Context, p *model.TimetablePlan) error
	DeletePlan(ctx context.Context, id string) error

	// time slot definitions
	CreateTimeSlot(ctx context.Context, ts *model.TimeSlotDefinition) error
	GetTimeSlot(ctx context.Context, id string) (*model.TimeSlotDefinition, error)
	ListTimeSlotsByPlan(ctx context.Context, planID string) ([]*model.TimeSlotDefinition, error)
	UpdateTimeSlot(ctx context.Context, ts *model.TimeSlotDefinition) error
	DeleteTimeSlot(ctx context.Context, id string) error
	DeleteTimeSlotsByPlan(ctx context.Context, planID string) error

	// grade requirements
	CreateRequirement(ctx context.Context, r *model.GradeRequirement) error
	GetRequirement(ctx context.Context, id string) (*model.GradeRequirement, error)
	ListRequirementsByPlan(ctx context.Context, planID string) ([]*model.GradeRequirement, error)
	UpdateRequirement(ctx context.Context, r *model.GradeRequirement) error
	DeleteRequirement(ctx context.Context, id string) error

	// class constraints
	CreateConstraint(ctx context.Context, c *model.ClassConstraint) error
	GetConstraint(ctx context.Context, id string) (*model.ClassConstraint, error)
	ListConstraintsByPlan(ctx context.Context, planID string) ([]*model.ClassConstraint, error)
	UpdateConstraint(ctx context.Context, c *model.ClassConstraint) error
	DeleteConstraint(ctx context.Context, id string) error

	// timetable entries
	CreateEntry(ctx context.Context, e *model.TimetableEntry) error
	BulkCreateEntries(ctx context.Context, entries []*model.TimetableEntry) error
	GetEntry(ctx context.Context, id string) (*model.TimetableEntry, error)
	ListEntriesByPlan(ctx context.Context, planID string) ([]*model.TimetableEntry, error)
	UpdateEntry(ctx context.Context, e *model.TimetableEntry) error
	DeleteEntry(ctx context.Context, id string) error
	DeleteEntriesByPlan(ctx context.Context, planID string) error

	// conflicts
	CreateConflict(ctx context.Context, c *model.Conflict) error
	GetConflict(ctx context.Context, id string) (*model.Conflict, error)
	ListConflictsByPlan(ctx context.Context, planID string) ([]*model.Conflict, error)
	DeleteConflict(ctx context.Context, id string) error
	DeleteConflictsByPlan(ctx context.Context, planID string) error

	// teacher snapshots
	CreateTeacherSnapshot(ctx context.Context, s *model.TeacherSnapshot) error
	ListTeacherSnapshots(ctx context.Context, planID string) ([]*model.TeacherSnapshot, error)

	// subject snapshots
	CreateSubjectSnapshot(ctx context.Context, s *model.SubjectSnapshot) error
	ListSubjectSnapshots(ctx context.Context, planID string) ([]*model.SubjectSnapshot, error)

	// school class snapshots
	CreateSchoolClassSnapshot(ctx context.Context, s *model.SchoolClassSnapshot) error
	ListSchoolClassSnapshots(ctx context.Context, planID string) ([]*model.SchoolClassSnapshot, error)

	// room snapshots
	CreateRoomSnapshot(ctx context.Context, s *model.RoomSnapshot) error
	ListRoomSnapshots(ctx context.Context, planID string) ([]*model.RoomSnapshot, error)

	// delete all snapshots for a plan
	DeleteSnapshotsByPlan(ctx context.Context, planID string) error
}
