//go:build integration

package integration_test

import (
	"fmt"
	"testing"

	"github.com/Joel-Haeberli/timetable-service/internal/client"
)

// TestConflict_TeacherDoubleBooked verifies that when two classes each need one
// Math lesson but only one time slot exists and only one teacher qualifies,
// a teacher_double_booked conflict is reported.
func TestConflict_TeacherDoubleBooked(t *testing.T) {
	planID, slotID, _ := setupDoubleBookedPlan(t)
	conflicts := listConflicts(t, planID)

	c := assertConflictType(t, conflicts, "teacher_double_booked")
	if c.TeacherID != teacherAlphaID {
		t.Fatalf("expected teacherId=%s, got %q", teacherAlphaID, c.TeacherID)
	}
	if c.TimeSlotID != slotID {
		t.Fatalf("expected timeSlotId=%s, got %q", slotID, c.TimeSlotID)
	}
	if len(c.EntryIDs) != 2 {
		t.Fatalf("expected 2 entryIds, got %d", len(c.EntryIDs))
	}
	if c.Severity != "error" {
		t.Fatalf("expected severity=error, got %q", c.Severity)
	}
}

// TestConflict_RoomDoubleBooked verifies that two entries sharing the same room
// and time slot produce a room_double_booked conflict.
func TestConflict_RoomDoubleBooked(t *testing.T) {
	// Use the double-booked plan (2 entries on same slot, plan=resolving)
	planID, slotID, entries := setupDoubleBookedPlan(t)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Assign the same room to both entries (they are already on the same slot)
	for _, e := range entries {
		mustDo(t, "PUT", fmt.Sprintf("/plans/%s/entries/%s", planID, e.ID), writeToken, map[string]any{
			"schoolClassId": e.SchoolClassID,
			"subjectId":     e.SubjectID,
			"subjectName":   e.SubjectName,
			"teacherId":     e.TeacherID,
			"roomId":        room101ID,
			"timeSlotId":    slotID,
			"version":       e.Version,
		}, 200, nil)
	}

	doValidate(t, planID)
	conflicts := listConflicts(t, planID)
	c := assertConflictType(t, conflicts, "room_double_booked")
	if c.TimeSlotID != slotID {
		t.Fatalf("expected timeSlotId=%s, got %q", slotID, c.TimeSlotID)
	}
}

// TestConflict_TeacherNotQualified verifies that reassigning an entry to a
// teacher who doesn't teach that subject produces a teacher_not_qualified conflict.
func TestConflict_TeacherNotQualified(t *testing.T) {
	// Set up a resolving plan using teacher_double_booked as the vehicle
	// (1 slot, 2 classes need Math, only TeacherAlpha qualifies) — both TeacherAlpha
	// and TeacherBeta are in the snapshot so reassignment to Beta is valid.
	p := newPlan(t)
	newSlot(t, p.ID, 1, 1, "08:00", "08:45", true)
	newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 1, 0, false)
	newRequirement(t, p.ID, class2bID, subjectMathID, subjectMathID, 1, 0, false)
	newConstraint(t, p.ID, class1aID, "1A", 5, 4, 3, "", true)
	newConstraint(t, p.ID, class2bID, "2B", 5, 4, 3, "", true)

	// TeacherBeta is in the snapshot (teaches History, not Math)
	alpha := client.TeacherData{ID: teacherAlphaID, Name: "Alpha", Prename: "T"}
	beta := client.TeacherData{ID: teacherBetaID, Name: "Beta", Prename: "T"}
	setMock(lsMockData{
		Teachers: []client.TeacherData{alpha, beta},
		Subjects: []client.SubjectData{
			{ID: subjectMathID, Name: subjectMathID, Teachers: []client.TeacherData{alpha}},
			{ID: subjectHistoryID, Name: subjectHistoryID, Teachers: []client.TeacherData{beta}},
		},
		Classes: []client.SchoolClassData{
			{ID: class1aID, Name: "1A"},
			{ID: class2bID, Name: "2B"},
		},
		Rooms: []client.RoomData{{ID: room101ID, Name: "Room 101"}},
	})
	doSnapshot(t, p.ID)
	doGenerate(t, p.ID)

	if status := getPlan(t, p.ID).Status; status != "resolving" {
		t.Fatalf("expected resolving, got %q", status)
	}

	// Find the class-1a Math entry and reassign it to TeacherBeta (unqualified for Math)
	entries := listEntries(t, p.ID)
	var mathEntry1a *entryResp
	for i := range entries {
		if entries[i].SchoolClassID == class1aID && entries[i].SubjectID == subjectMathID {
			mathEntry1a = &entries[i]
			break
		}
	}
	if mathEntry1a == nil {
		t.Fatal("could not find class-1a Math entry")
	}

	mustDo(t, "POST", fmt.Sprintf("/plans/%s/entries/%s/reassign", p.ID, mathEntry1a.ID), writeToken, map[string]any{
		"teacherId": teacherBetaID,
	}, 200, nil)

	doValidate(t, p.ID)
	conflicts := listConflicts(t, p.ID)
	c := assertConflictType(t, conflicts, "teacher_not_qualified")
	if c.TeacherID != teacherBetaID {
		t.Fatalf("expected teacherId=%s, got %q", teacherBetaID, c.TeacherID)
	}
	if c.SchoolClassID != class1aID {
		t.Fatalf("expected schoolClassId=%s, got %q", class1aID, c.SchoolClassID)
	}
}

// TestConflict_TooManyLessons verifies that when a class needs more lessons
// than available slots, a too_many_lessons conflict is reported.
func TestConflict_TooManyLessons(t *testing.T) {
	p := newPlan(t)
	// Only 2 slots but 3 lessons required
	newSlot(t, p.ID, 1, 1, "08:00", "08:45", true)
	newSlot(t, p.ID, 2, 1, "08:00", "08:45", true)
	newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 3, 0, false)
	newConstraint(t, p.ID, class1aID, "1A", 5, 4, 3, "", true)
	setMock(alphaOnlyMathSnapshot())
	doSnapshot(t, p.ID)
	doGenerate(t, p.ID)

	conflicts := listConflicts(t, p.ID)
	c := assertConflictType(t, conflicts, "too_many_lessons")
	if c.SchoolClassID != class1aID {
		t.Fatalf("expected schoolClassId=%s, got %q", class1aID, c.SchoolClassID)
	}
	// Description should mention both actual count and required count
	if c.Description == "" {
		t.Fatal("expected non-empty conflict description")
	}
}

// TestConflict_MaxEarlyStartsExceeded verifies that placing more period-1 lessons
// than maxEarlyStarts allows produces a max_early_starts_exceeded conflict.
func TestConflict_MaxEarlyStartsExceeded(t *testing.T) {
	p := newPlan(t)
	// All 3 slots are period=1 on different days
	newSlot(t, p.ID, 1, 1, "08:00", "08:45", true) // Mon period 1
	newSlot(t, p.ID, 2, 1, "08:00", "08:45", true) // Tue period 1
	newSlot(t, p.ID, 3, 1, "08:00", "08:45", true) // Wed period 1
	// Require 3 Math lessons, maxEarlyStarts=0 → all period-1 placements violate the constraint
	newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 3, 0, false)
	newConstraint(t, p.ID, class1aID, "1A", 0, 4, 3, "", true) // maxEarlyStarts=0
	setMock(alphaOnlyMathSnapshot())
	doSnapshot(t, p.ID)
	doGenerate(t, p.ID)

	conflicts := listConflicts(t, p.ID)
	c := assertConflictType(t, conflicts, "max_early_starts_exceeded")
	if c.SchoolClassID != class1aID {
		t.Fatalf("expected schoolClassId=%s, got %q", class1aID, c.SchoolClassID)
	}
}

// TestConflict_FreeAfternoonViolated verifies that a class with an afternoon
// lesson on its designated free afternoon day produces a free_afternoon_violated
// conflict.
func TestConflict_FreeAfternoonViolated(t *testing.T) {
	p := newPlan(t)
	// 4 morning slots (Mon-Thu period 1) + 1 afternoon slot (Wed period 5)
	newSlot(t, p.ID, 1, 1, "08:00", "08:45", true) // Mon morning
	newSlot(t, p.ID, 2, 1, "08:00", "08:45", true) // Tue morning
	newSlot(t, p.ID, 3, 1, "08:00", "08:45", true) // Wed morning
	newSlot(t, p.ID, 4, 1, "08:00", "08:45", true) // Thu morning
	newSlot(t, p.ID, 3, 5, "13:00", "13:45", false) // Wed afternoon (period 5, not morning)
	// Require 5 Math lessons — 4 morning slots used, 5th falls back to Wed afternoon
	newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 5, 0, false)
	// Free afternoons on Wednesday (day 3)
	newConstraint(t, p.ID, class1aID, "1A", 5, 4, 3, "3", true)
	setMock(alphaOnlyMathSnapshot())
	doSnapshot(t, p.ID)
	doGenerate(t, p.ID)

	conflicts := listConflicts(t, p.ID)
	c := assertConflictType(t, conflicts, "free_afternoon_violated")
	if c.SchoolClassID != class1aID {
		t.Fatalf("expected schoolClassId=%s, got %q", class1aID, c.SchoolClassID)
	}
	// Description should mention Wednesday
	if c.Description == "" {
		t.Fatal("expected non-empty description")
	}
}

// TestConflict_MaxDoubleLessonsExceeded verifies that updating a requirement to
// reduce maxDoubleLessons below the actual number of double-lesson pairs
// produces a max_double_lessons_exceeded conflict on the next validation.
func TestConflict_MaxDoubleLessonsExceeded(t *testing.T) {
	p := newPlan(t)
	// 4 consecutive slots: Mon p1+p2, Tue p1+p2
	newSlot(t, p.ID, 1, 1, "08:00", "08:45", true)
	newSlot(t, p.ID, 1, 2, "08:50", "09:35", true)
	newSlot(t, p.ID, 2, 1, "08:00", "08:45", true)
	newSlot(t, p.ID, 2, 2, "08:50", "09:35", true)

	// Generate with maxDoubleLessons=1: planner creates 1 double pair + 2 singles
	req := newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 4, 1, false)
	newConstraint(t, p.ID, class1aID, "1A", 5, 4, 3, "", true)
	setMock(alphaOnlyMathSnapshot())
	doSnapshot(t, p.ID)
	g := doGenerate(t, p.ID)

	// Should have 4 entries with no conflicts
	if g.Entries != 4 {
		t.Fatalf("expected 4 entries, got %d", g.Entries)
	}
	if g.Conflicts != 0 {
		t.Fatalf("expected no conflicts initially, got %d", g.Conflicts)
	}

	// Now tighten the constraint: reduce maxDoubleLessons to 0
	mustDo(t, "PUT", fmt.Sprintf("/plans/%s/requirements/%s", p.ID, req.ID), writeToken, map[string]any{
		"schoolClassId":    class1aID,
		"subjectId":        subjectMathID,
		"subjectName":      subjectMathID,
		"lessonsPerWeek":   4,
		"maxDoubleLessons": 0, // now 0, but 1 pair exists → conflict
		"lessonDurationMin": 45,
		"version":          req.Version,
	}, 200, nil)

	v := doValidate(t, p.ID)
	c := assertConflictType(t, v.Items, "max_double_lessons_exceeded")
	if c.SchoolClassID != class1aID {
		t.Fatalf("expected schoolClassId=%s, got %q", class1aID, c.SchoolClassID)
	}
}

// TestConflict_MultipleSimultaneous verifies that a plan can generate multiple
// distinct conflict types at the same time.
func TestConflict_MultipleSimultaneous(t *testing.T) {
	p := newPlan(t)
	// Single period-1 slot, maxEarlyStarts=0 for both classes
	newSlot(t, p.ID, 1, 1, "08:00", "08:45", true)
	newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 1, 0, false)
	newRequirement(t, p.ID, class2bID, subjectMathID, subjectMathID, 1, 0, false)
	newConstraint(t, p.ID, class1aID, "1A", 0, 4, 3, "", true) // maxEarlyStarts=0
	newConstraint(t, p.ID, class2bID, "2B", 0, 4, 3, "", true) // maxEarlyStarts=0
	setMock(alphaOnlyMathSnapshot())
	doSnapshot(t, p.ID)
	doGenerate(t, p.ID)

	conflicts := listConflicts(t, p.ID)
	if len(conflicts) < 2 {
		t.Fatalf("expected at least 2 conflicts, got %d: %v", len(conflicts), conflictTypes(conflicts))
	}

	// Must have teacher_double_booked (both classes same slot, same teacher)
	assertConflictType(t, conflicts, "teacher_double_booked")
	// Must have max_early_starts_exceeded (period=1 slots used, maxEarlyStarts=0)
	assertConflictType(t, conflicts, "max_early_starts_exceeded")

	// Verify distinct types
	types := conflictTypes(conflicts)
	seen := map[string]bool{}
	for _, ct := range types {
		seen[ct] = true
	}
	if len(seen) < 2 {
		t.Fatalf("expected at least 2 distinct conflict types, got: %v", types)
	}
}

func conflictTypes(conflicts []conflictResp) []string {
	types := make([]string, len(conflicts))
	for i, c := range conflicts {
		types[i] = c.Type
	}
	return types
}
