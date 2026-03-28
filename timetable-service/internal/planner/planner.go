package planner

import (
	"context"
	"fmt"

	"github.com/Joel-Haeberli/timetable-service/internal/model"
	"github.com/Joel-Haeberli/timetable-service/internal/storage"
)

// Planner orchestrates the timetable generation algorithm.
type Planner struct {
	store storage.Storage
}

// New creates a Planner backed by the given storage.
func New(store storage.Storage) *Planner {
	return &Planner{store: store}
}

// Generate runs the full planning pipeline and returns counts of entries
// created and conflicts found.
func (p *Planner) Generate(ctx context.Context, planID string) (int, int, error) {
	// Load all data needed for generation
	requirements, err := p.store.ListRequirementsByPlan(ctx, planID)
	if err != nil {
		return 0, 0, fmt.Errorf("loading requirements: %w", err)
	}
	constraints, err := p.store.ListConstraintsByPlan(ctx, planID)
	if err != nil {
		return 0, 0, fmt.Errorf("loading constraints: %w", err)
	}
	timeSlots, err := p.store.ListTimeSlotsByPlan(ctx, planID)
	if err != nil {
		return 0, 0, fmt.Errorf("loading time slots: %w", err)
	}
	teachers, err := p.store.ListTeacherSnapshots(ctx, planID)
	if err != nil {
		return 0, 0, fmt.Errorf("loading teacher snapshots: %w", err)
	}
	classes, err := p.store.ListSchoolClassSnapshots(ctx, planID)
	if err != nil {
		return 0, 0, fmt.Errorf("loading class snapshots: %w", err)
	}
	rooms, err := p.store.ListRoomSnapshots(ctx, planID)
	if err != nil {
		return 0, 0, fmt.Errorf("loading room snapshots: %w", err)
	}

	// Clear existing entries and conflicts
	if err := p.store.DeleteConflictsByPlan(ctx, planID); err != nil {
		return 0, 0, fmt.Errorf("clearing conflicts: %w", err)
	}
	if err := p.store.DeleteEntriesByPlan(ctx, planID); err != nil {
		return 0, 0, fmt.Errorf("clearing entries: %w", err)
	}

	// Phase 2: Allocate time slots per class
	entries := AllocateSlots(requirements, constraints, timeSlots)
	for _, e := range entries {
		e.PlanID = planID
	}

	// Phase 3: Assign teachers
	AssignTeachers(entries, teachers, timeSlots)

	// Phase 4: Assign rooms
	AssignRooms(entries, classes, rooms)

	// Persist entries
	if err := p.store.BulkCreateEntries(ctx, entries); err != nil {
		return 0, 0, fmt.Errorf("persisting entries: %w", err)
	}

	// Phase 5: Detect conflicts
	conflicts := Validate(entries, requirements, constraints, teachers, timeSlots)
	for _, c := range conflicts {
		c.PlanID = planID
		if err := p.store.CreateConflict(ctx, c); err != nil {
			return 0, 0, fmt.Errorf("persisting conflict: %w", err)
		}
	}

	return len(entries), len(conflicts), nil
}

// ValidateOnly runs conflict detection on existing entries without regenerating.
func (p *Planner) ValidateOnly(ctx context.Context, planID string) ([]*model.Conflict, error) {
	entries, err := p.store.ListEntriesByPlan(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("loading entries: %w", err)
	}
	requirements, err := p.store.ListRequirementsByPlan(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("loading requirements: %w", err)
	}
	constraints, err := p.store.ListConstraintsByPlan(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("loading constraints: %w", err)
	}
	teachers, err := p.store.ListTeacherSnapshots(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("loading teacher snapshots: %w", err)
	}
	timeSlots, err := p.store.ListTimeSlotsByPlan(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("loading time slots: %w", err)
	}

	// Clear old conflicts
	if err := p.store.DeleteConflictsByPlan(ctx, planID); err != nil {
		return nil, fmt.Errorf("clearing conflicts: %w", err)
	}

	conflicts := Validate(entries, requirements, constraints, teachers, timeSlots)
	for _, c := range conflicts {
		c.PlanID = planID
		if err := p.store.CreateConflict(ctx, c); err != nil {
			return nil, fmt.Errorf("persisting conflict: %w", err)
		}
	}

	return conflicts, nil
}
