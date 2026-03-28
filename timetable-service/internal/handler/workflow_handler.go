package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Joel-Haeberli/timetable-service/internal/client"
	"github.com/Joel-Haeberli/timetable-service/internal/model"
	"github.com/Joel-Haeberli/timetable-service/internal/planner"
	"github.com/Joel-Haeberli/timetable-service/internal/storage"
)

// WorkflowHandler manages the timetable planning lifecycle operations.
type WorkflowHandler struct {
	store   storage.Storage
	client  *client.LeanschoolClient
	planner *planner.Planner
}

func NewWorkflowHandler(store storage.Storage, lsClient *client.LeanschoolClient, p *planner.Planner) *WorkflowHandler {
	return &WorkflowHandler{store: store, client: lsClient, planner: p}
}

func (h *WorkflowHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /plans/{planId}/snapshot", h.Snapshot)
	mux.HandleFunc("POST /plans/{planId}/generate", h.Generate)
	mux.HandleFunc("POST /plans/{planId}/validate", h.Validate)
	mux.HandleFunc("POST /plans/{planId}/finalize", h.Finalize)
	mux.HandleFunc("POST /plans/{planId}/reset", h.Reset)
}

// snapshotSummary is the response body for a successful snapshot operation.
type snapshotSummary struct {
	Teachers int `json:"teachers"`
	Subjects int `json:"subjects"`
	Classes  int `json:"classes"`
	Rooms    int `json:"rooms"`
}

// Snapshot captures current teachers, subjects, classes, and rooms from the
// leanschool API into plan-local snapshot tables.
func (h *WorkflowHandler) Snapshot(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	planID := r.PathValue("planId")

	// Verify the plan exists and is in draft status.
	plan, err := h.store.GetPlan(r.Context(), planID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if plan == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if plan.Status != model.PlanStatusDraft {
		http.Error(w, "plan must be in draft status to take a snapshot", http.StatusConflict)
		return
	}

	// Fetch master data from the leanschool API.
	teachers, err := h.client.FetchTeachers()
	if err != nil {
		log.Printf("[snapshot] fetching teachers: %v", err)
		http.Error(w, "failed to fetch teachers from leanschool", http.StatusBadGateway)
		return
	}

	subjects, err := h.client.FetchSubjects()
	if err != nil {
		log.Printf("[snapshot] fetching subjects: %v", err)
		http.Error(w, "failed to fetch subjects from leanschool", http.StatusBadGateway)
		return
	}

	classes, err := h.client.FetchSchoolClasses()
	if err != nil {
		log.Printf("[snapshot] fetching school classes: %v", err)
		http.Error(w, "failed to fetch school classes from leanschool", http.StatusBadGateway)
		return
	}

	rooms, err := h.client.FetchRooms()
	if err != nil {
		log.Printf("[snapshot] fetching rooms: %v", err)
		http.Error(w, "failed to fetch rooms from leanschool", http.StatusBadGateway)
		return
	}

	// Delete existing snapshots for this plan (idempotent refresh).
	if err := h.store.DeleteSnapshotsByPlan(r.Context(), planID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build a lookup of teacher ID → subject names from the subjects response
	// (subjects carry their assigned teachers).
	teacherSubjects := make(map[string][]string)
	for _, s := range subjects {
		for _, t := range s.Teachers {
			teacherSubjects[t.ID] = append(teacherSubjects[t.ID], s.Name)
		}
	}

	// Store teacher snapshots.
	for _, t := range teachers {
		snap := &model.TeacherSnapshot{
			ID:       t.ID,
			PlanID:   planID,
			Name:     t.Name,
			Prename:  t.Prename,
			Sub:      t.Sub,
			Subjects: teacherSubjects[t.ID],
		}
		if snap.Subjects == nil {
			snap.Subjects = []string{}
		}
		if err := h.store.CreateTeacherSnapshot(r.Context(), snap); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Store subject snapshots.
	for _, s := range subjects {
		snap := &model.SubjectSnapshot{
			ID:     s.ID,
			PlanID: planID,
			Name:   s.Name,
		}
		if err := h.store.CreateSubjectSnapshot(r.Context(), snap); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Store school class snapshots.
	for _, c := range classes {
		snap := &model.SchoolClassSnapshot{
			ID:       c.ID,
			PlanID:   planID,
			Name:     c.Name,
			Shortcut: c.Shortcut,
		}
		if err := h.store.CreateSchoolClassSnapshot(r.Context(), snap); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Store room snapshots.
	for _, rm := range rooms {
		snap := &model.RoomSnapshot{
			ID:       rm.ID,
			PlanID:   planID,
			Name:     rm.Name,
			RoomType: rm.RoomType,
		}
		if err := h.store.CreateRoomSnapshot(r.Context(), snap); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshotSummary{
		Teachers: len(teachers),
		Subjects: len(subjects),
		Classes:  len(classes),
		Rooms:    len(rooms),
	})
}

// generateResponse is the response body for a successful generate operation.
type generateResponse struct {
	Entries   int `json:"entries"`
	Conflicts int `json:"conflicts"`
}

// Generate runs the planning algorithm to create timetable entries from
// requirements, constraints, and snapshots.
func (h *WorkflowHandler) Generate(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	planID := r.PathValue("planId")

	plan, err := h.store.GetPlan(r.Context(), planID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if plan == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if plan.Status != model.PlanStatusDraft && plan.Status != model.PlanStatusPlanning {
		http.Error(w, "plan must be in draft or planning status to generate", http.StatusConflict)
		return
	}

	numEntries, numConflicts, err := h.planner.Generate(r.Context(), planID)
	if err != nil {
		log.Printf("[generate] error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Transition status based on conflicts.
	if numConflicts > 0 {
		plan.Status = model.PlanStatusResolving
	} else {
		plan.Status = model.PlanStatusAccepted
	}
	if err := h.store.UpdatePlan(r.Context(), plan); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(generateResponse{
		Entries:   numEntries,
		Conflicts: numConflicts,
	})
}

// validateResponse is the response body for a successful validate operation.
type validateResponse struct {
	Conflicts int              `json:"conflicts"`
	Items     []*model.Conflict `json:"items"`
}

// Validate detects scheduling conflicts in the current timetable entries.
func (h *WorkflowHandler) Validate(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	planID := r.PathValue("planId")

	plan, err := h.store.GetPlan(r.Context(), planID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if plan == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	conflicts, err := h.planner.ValidateOnly(r.Context(), planID)
	if err != nil {
		log.Printf("[validate] error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update status based on conflicts if plan is in resolving or planning.
	if plan.Status == model.PlanStatusResolving || plan.Status == model.PlanStatusPlanning {
		if len(conflicts) > 0 {
			plan.Status = model.PlanStatusResolving
		} else {
			plan.Status = model.PlanStatusAccepted
		}
		if err := h.store.UpdatePlan(r.Context(), plan); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(validateResponse{
		Conflicts: len(conflicts),
		Items:     conflicts,
	})
}

// finalizeResponse is the response body for a successful finalize operation.
type finalizeResponse struct {
	LessonsCreated int `json:"lessonsCreated"`
}

// Finalize accepts the timetable and writes lessons to the leanschool API.
func (h *WorkflowHandler) Finalize(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	planID := r.PathValue("planId")

	plan, err := h.store.GetPlan(r.Context(), planID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if plan == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if plan.Status != model.PlanStatusAccepted {
		http.Error(w, "plan must be in accepted status to finalize", http.StatusConflict)
		return
	}

	// Load entries and time slots to build lessons.
	entries, err := h.store.ListEntriesByPlan(r.Context(), planID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeSlots, err := h.store.ListTimeSlotsByPlan(r.Context(), planID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tsMap := make(map[string]*model.TimeSlotDefinition, len(timeSlots))
	for _, ts := range timeSlots {
		tsMap[ts.ID] = ts
	}

	created := 0
	for _, e := range entries {
		ts := tsMap[e.TimeSlotID]
		dow := int(ts.DayOfWeek)
		period := ts.Period

		lesson := client.LessonData{
			Teacher:     &client.EntityRef{ID: e.TeacherID},
			SchoolClass: &client.EntityRef{ID: e.SchoolClassID},
			Subject:     &client.EntityRef{ID: e.SubjectID},
			DayOfWeek:   &dow,
			Period:      &period,
			StartTime:   ts.StartTime,
			EndTime:     ts.EndTime,
		}
		if e.RoomID != "" {
			lesson.Room = &client.EntityRef{ID: e.RoomID}
		}

		if _, err := h.client.CreateLesson(lesson); err != nil {
			log.Printf("[finalize] creating lesson for entry %s: %v", e.ID, err)
			http.Error(w, fmt.Sprintf("failed to create lesson: %v", err), http.StatusBadGateway)
			return
		}
		created++
	}

	// Mark plan as finalized.
	plan.Status = model.PlanStatusFinalized
	if err := h.store.UpdatePlan(r.Context(), plan); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(finalizeResponse{LessonsCreated: created})
}

// Reset clears all entries, conflicts, and snapshots for a plan, setting its
// status back to "draft". This is fully implemented.
func (h *WorkflowHandler) Reset(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	planID := r.PathValue("planId")

	plan, err := h.store.GetPlan(r.Context(), planID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if plan == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Delete conflicts first (they reference entries via join table)
	if err := h.store.DeleteConflictsByPlan(r.Context(), planID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.store.DeleteEntriesByPlan(r.Context(), planID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.store.DeleteSnapshotsByPlan(r.Context(), planID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reset plan status to draft
	plan.Status = model.PlanStatusDraft
	if err := h.store.UpdatePlan(r.Context(), plan); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}
