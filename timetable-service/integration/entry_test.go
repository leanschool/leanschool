//go:build integration

package integration_test

import (
	"fmt"
	"testing"
)

// setupResolvingPlan creates a resolving plan with 2 entries on different slots.
// It uses class-1a, Math 3 lessons, 2 time slots → too_many_lessons conflict.
// Returns (planID, slotAID, slotBID).
func setupResolvingPlan(t *testing.T) (planID, slotA, slotB string) {
	t.Helper()
	p := newPlan(t)
	planID = p.ID

	// 2 slots on different days/periods
	sA := newSlot(t, planID, 1, 1, "08:00", "08:45", true)
	sB := newSlot(t, planID, 2, 1, "08:00", "08:45", true)
	slotA = sA.ID
	slotB = sB.ID

	newRequirement(t, planID, class1aID, subjectMathID, subjectMathID, 3, 0, false)
	newConstraint(t, planID, class1aID, "1A", 5, 4, 3, "", true)

	setMock(alphaOnlyMathSnapshot())
	doSnapshot(t, planID)

	g := doGenerate(t, planID)
	// With 3 lessons required but only 2 slots, planner creates 2 entries → too_many_lessons
	if g.Conflicts == 0 {
		t.Fatalf("expected conflicts from insufficient slots, got none (entries=%d)", g.Entries)
	}

	if status := getPlan(t, planID).Status; status != "resolving" {
		t.Fatalf("expected status=resolving, got %q", status)
	}
	return
}

// setupDoubleBookedPlan creates a resolving plan where both class-1a and class-2b
// need 1 Math lesson but only 1 time slot is available, causing teacher_double_booked.
// Returns (planID, slotID, entry1ID, entry2ID).
func setupDoubleBookedPlan(t *testing.T) (planID, slotID string, entries []entryResp) {
	t.Helper()
	p := newPlan(t)
	planID = p.ID

	s := newSlot(t, planID, 1, 1, "08:00", "08:45", true)
	slotID = s.ID

	newRequirement(t, planID, class1aID, subjectMathID, subjectMathID, 1, 0, false)
	newRequirement(t, planID, class2bID, subjectMathID, subjectMathID, 1, 0, false)
	newConstraint(t, planID, class1aID, "1A", 5, 4, 3, "", true)
	newConstraint(t, planID, class2bID, "2B", 5, 4, 3, "", true)

	setMock(alphaOnlyMathSnapshot())
	doSnapshot(t, planID)
	doGenerate(t, planID)

	entries = listEntries(t, planID)
	return
}

func TestEntry_List(t *testing.T) {
	planID, _, _ := setupResolvingPlan(t)
	entries := listEntries(t, planID)
	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry")
	}
}

func TestEntry_Get(t *testing.T) {
	planID, _, _ := setupResolvingPlan(t)
	entries := listEntries(t, planID)
	if len(entries) == 0 {
		t.Skip("no entries to get")
	}

	var got entryResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/entries/%s", planID, entries[0].ID), readToken, nil, 200, &got)
	if got.ID != entries[0].ID {
		t.Fatalf("expected id=%s, got %s", entries[0].ID, got.ID)
	}
}

func TestEntry_UpdateRequiresResolvingStatus(t *testing.T) {
	// Use a clean plan in draft status (no entries, can't test with entries)
	// Instead, create an accepted plan and try to modify entries.
	p := newPlan(t)

	// Set up a plan that will have no conflicts (1 class, adequate slots, qualified teacher)
	generateDefaultSlots(t, p.ID)
	newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 1, 0, false)
	newConstraint(t, p.ID, class1aID, "1A", 5, 4, 3, "", true)
	setMock(alphaOnlyMathSnapshot())
	doSnapshot(t, p.ID)
	doGenerate(t, p.ID)

	if status := getPlan(t, p.ID).Status; status != "accepted" {
		t.Fatalf("expected accepted plan, got %q", status)
	}

	entries := listEntries(t, p.ID)
	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry")
	}
	e := entries[0]

	// PUT on accepted plan → 400 (not resolving)
	mustDo(t, "PUT", fmt.Sprintf("/plans/%s/entries/%s", p.ID, e.ID), writeToken, map[string]any{
		"schoolClassId": e.SchoolClassID,
		"subjectId":     e.SubjectID,
		"teacherId":     e.TeacherID,
		"timeSlotId":    e.TimeSlotID,
		"version":       e.Version,
	}, 400, nil)
}

func TestEntry_Update_HappyPath(t *testing.T) {
	planID, slotA, slotB := setupResolvingPlan(t)
	entries := listEntries(t, planID)
	if len(entries) == 0 {
		t.Fatal("no entries")
	}
	e := entries[0]
	// Move entry to the other slot
	newSlot := slotB
	if e.TimeSlotID == slotB {
		newSlot = slotA
	}

	var updated entryResp
	mustDo(t, "PUT", fmt.Sprintf("/plans/%s/entries/%s", planID, e.ID), writeToken, map[string]any{
		"schoolClassId": e.SchoolClassID,
		"subjectId":     e.SubjectID,
		"subjectName":   e.SubjectName,
		"teacherId":     e.TeacherID,
		"timeSlotId":    newSlot,
		"version":       e.Version,
	}, 200, &updated)

	if updated.TimeSlotID != newSlot {
		t.Fatalf("expected timeSlotId=%s, got %s", newSlot, updated.TimeSlotID)
	}
	if updated.Version != e.Version+1 {
		t.Fatalf("expected version=%d, got %d", e.Version+1, updated.Version)
	}
}

func TestEntry_Update_ForbiddenWithReadToken(t *testing.T) {
	planID, slotA, _ := setupResolvingPlan(t)
	entries := listEntries(t, planID)
	if len(entries) == 0 {
		t.Skip("no entries")
	}
	e := entries[0]

	mustDo(t, "PUT", fmt.Sprintf("/plans/%s/entries/%s", planID, e.ID), readToken, map[string]any{
		"schoolClassId": e.SchoolClassID,
		"subjectId":     e.SubjectID,
		"timeSlotId":    slotA,
		"version":       e.Version,
	}, 403, nil)
}

func TestEntry_Update_AllowedWithResolveToken(t *testing.T) {
	planID, slotA, slotB := setupResolvingPlan(t)
	entries := listEntries(t, planID)
	if len(entries) == 0 {
		t.Skip("no entries")
	}
	e := entries[0]
	newSlot := slotB
	if e.TimeSlotID == slotB {
		newSlot = slotA
	}

	mustDo(t, "PUT", fmt.Sprintf("/plans/%s/entries/%s", planID, e.ID), resolveToken, map[string]any{
		"schoolClassId": e.SchoolClassID,
		"subjectId":     e.SubjectID,
		"subjectName":   e.SubjectName,
		"teacherId":     e.TeacherID,
		"timeSlotId":    newSlot,
		"version":       e.Version,
	}, 200, nil)
}

func TestEntry_Swap(t *testing.T) {
	planID, slotA, slotB := setupResolvingPlan(t)
	entries := listEntries(t, planID)
	if len(entries) < 2 {
		t.Skip("need at least 2 entries for swap")
	}

	// Find two entries on different slots
	var e1, e2 *entryResp
	for i := range entries {
		for j := range entries {
			if entries[i].TimeSlotID != entries[j].TimeSlotID {
				e1 = &entries[i]
				e2 = &entries[j]
				break
			}
		}
		if e1 != nil {
			break
		}
	}
	if e1 == nil {
		t.Skip("no two entries on different slots")
	}
	_ = slotA
	_ = slotB

	origSlot1 := e1.TimeSlotID
	origSlot2 := e2.TimeSlotID

	type swapResp struct {
		Source entryResp `json:"source"`
		Target entryResp `json:"target"`
	}
	var result swapResp
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/entries/%s/swap", planID, e1.ID), writeToken, map[string]any{
		"targetEntryId": e2.ID,
	}, 200, &result)

	if result.Source.TimeSlotID != origSlot2 {
		t.Fatalf("source should have target's original slot; got %s, want %s", result.Source.TimeSlotID, origSlot2)
	}
	if result.Target.TimeSlotID != origSlot1 {
		t.Fatalf("target should have source's original slot; got %s, want %s", result.Target.TimeSlotID, origSlot1)
	}
}

func TestEntry_Reassign(t *testing.T) {
	planID, _, _ := setupResolvingPlan(t)
	entries := listEntries(t, planID)
	if len(entries) == 0 {
		t.Skip("no entries")
	}
	e := entries[0]

	// Reassign to teacherBeta who doesn't teach Math
	var updated entryResp
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/entries/%s/reassign", planID, e.ID), writeToken, map[string]any{
		"teacherId": teacherBetaID,
	}, 200, &updated)

	if updated.TeacherID != teacherBetaID {
		t.Fatalf("expected teacherId=%s, got %s", teacherBetaID, updated.TeacherID)
	}

	// Validate to get conflict for unqualified teacher
	v := doValidate(t, planID)
	conflicts := make([]conflictResp, len(v.Items))
	copy(conflicts, v.Items)
	assertConflictType(t, conflicts, "teacher_not_qualified")
}

func TestEntry_FilterByClass(t *testing.T) {
	planID, _, _ := setupDoubleBookedPlan(t)

	var filtered []entryResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/entries?classId=%s", planID, class1aID), readToken, nil, 200, &filtered)
	for _, e := range filtered {
		if e.SchoolClassID != class1aID {
			t.Fatalf("filter returned entry for class %s, expected %s", e.SchoolClassID, class1aID)
		}
	}
	if len(filtered) == 0 {
		t.Fatal("expected at least 1 entry for class-1a")
	}
}

func TestEntry_FilterByTeacher(t *testing.T) {
	planID, _, _ := setupDoubleBookedPlan(t)

	var filtered []entryResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/entries?teacherId=%s", planID, teacherAlphaID), readToken, nil, 200, &filtered)
	for _, e := range filtered {
		if e.TeacherID != teacherAlphaID {
			t.Fatalf("filter returned entry for teacher %s, expected %s", e.TeacherID, teacherAlphaID)
		}
	}
	if len(filtered) == 0 {
		t.Fatal("expected at least 1 entry for teacherAlpha")
	}
}

func TestEntry_OptimisticLock(t *testing.T) {
	planID, slotA, slotB := setupResolvingPlan(t)
	entries := listEntries(t, planID)
	if len(entries) == 0 {
		t.Skip("no entries")
	}
	e := entries[0]
	newSlot := slotB
	if e.TimeSlotID == slotB {
		newSlot = slotA
	}

	body := map[string]any{
		"schoolClassId": e.SchoolClassID,
		"subjectId":     e.SubjectID,
		"subjectName":   e.SubjectName,
		"teacherId":     e.TeacherID,
		"timeSlotId":    newSlot,
		"version":       e.Version,
	}
	mustDo(t, "PUT", fmt.Sprintf("/plans/%s/entries/%s", planID, e.ID), writeToken, body, 200, nil)
	mustDo(t, "PUT", fmt.Sprintf("/plans/%s/entries/%s", planID, e.ID), writeToken, body, 409, nil) // stale
}
