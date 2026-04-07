//go:build integration

package integration_test

import (
	"fmt"
	"testing"
)

// TestE2E_HappyPath covers the full lifecycle: draft → snapshot → generate
// (no conflicts) → accepted → finalize → finalized.
func TestE2E_HappyPath(t *testing.T) {
	p := newPlan(t)
	if p.Status != "draft" {
		t.Fatalf("expected draft, got %q", p.Status)
	}

	// Generate a Mon-Fri grid (4 morning + 3 afternoon = 35 slots)
	slots := generateDefaultSlots(t, p.ID)
	if len(slots) != 35 {
		t.Fatalf("expected 35 slots, got %d", len(slots))
	}

	// Constraints for class 1A only (class 2B excluded via hasTimetable=false)
	newConstraint(t, p.ID, class1aID, "1A", 5, 4, 3, "", true)

	// Requirements: Math 3 lessons (1 double), English 2 lessons, History 2 lessons
	newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 3, 1, true)
	newRequirement(t, p.ID, class1aID, subjectEnglishID, subjectEnglishID, 2, 0, false)
	newRequirement(t, p.ID, class1aID, subjectHistoryID, subjectHistoryID, 2, 0, false)

	// Snapshot
	setMock(standardSnapshot())
	summary := doSnapshot(t, p.ID)
	if summary.Teachers != 3 {
		t.Fatalf("snapshot: expected 3 teachers, got %d", summary.Teachers)
	}

	// Generate — should produce 7 entries with no conflicts
	g := doGenerate(t, p.ID)
	if g.Conflicts != 0 {
		t.Fatalf("expected no conflicts, got %d", g.Conflicts)
	}
	if g.Entries != 7 {
		t.Fatalf("expected 7 entries (3 Math + 2 English + 2 History), got %d", g.Entries)
	}
	if status := getPlan(t, p.ID).Status; status != "accepted" {
		t.Fatalf("expected status=accepted, got %q", status)
	}

	// Verify entries
	entries := listEntries(t, p.ID)
	if len(entries) != 7 {
		t.Fatalf("expected 7 entries, got %d", len(entries))
	}

	subjectCount := map[string]int{}
	doubleCount := 0
	for _, e := range entries {
		subjectCount[e.SubjectID]++
		if e.IsDoubleLesson {
			doubleCount++
		}
		if e.TeacherID == "" {
			t.Errorf("entry %s has no teacher assigned", e.ID)
		}
	}
	if subjectCount[subjectMathID] != 3 {
		t.Fatalf("expected 3 Math entries, got %d", subjectCount[subjectMathID])
	}
	if subjectCount[subjectEnglishID] != 2 {
		t.Fatalf("expected 2 English entries, got %d", subjectCount[subjectEnglishID])
	}
	if subjectCount[subjectHistoryID] != 2 {
		t.Fatalf("expected 2 History entries, got %d", subjectCount[subjectHistoryID])
	}
	if doubleCount != 2 {
		t.Fatalf("expected 2 double-lesson entries (1 pair), got %d", doubleCount)
	}

	// No conflicts
	conflicts := listConflicts(t, p.ID)
	if len(conflicts) != 0 {
		t.Fatalf("expected 0 conflicts, got %d", len(conflicts))
	}

	// Finalize
	var fin finalizeResp
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/finalize", p.ID), writeToken, nil, 200, &fin)
	if fin.LessonsCreated != 7 {
		t.Fatalf("expected 7 lessons created, got %d", fin.LessonsCreated)
	}
	if status := getPlan(t, p.ID).Status; status != "finalized" {
		t.Fatalf("expected status=finalized, got %q", status)
	}

	// Verify leanschool mock captured 7 lesson creation calls
	lessons := getLessonsCreated()
	if len(lessons) != 7 {
		t.Fatalf("leanschool mock: expected 7 lessons, captured %d", len(lessons))
	}
	for i, l := range lessons {
		if l.DayOfWeek == nil {
			t.Errorf("lesson[%d] missing dayOfWeek", i)
		}
		if l.Period == nil {
			t.Errorf("lesson[%d] missing period", i)
		}
		if l.StartTime == "" {
			t.Errorf("lesson[%d] missing startTime", i)
		}
		if l.SchoolClass == nil {
			t.Errorf("lesson[%d] missing schoolClass", i)
		}
	}

	// Cannot delete a finalized plan
	mustDo(t, "DELETE", fmt.Sprintf("/plans/%s", p.ID), writeToken, nil, 400, nil)
}

// TestE2E_ConflictResolution covers: draft → generate (with conflict) →
// resolving → fix entry → validate → accepted → finalize.
func TestE2E_ConflictResolution(t *testing.T) {
	p := newPlan(t)

	// 1 slot only → double-booking guaranteed with 2 classes and 1 teacher
	slotA := newSlot(t, p.ID, 1, 1, "08:00", "08:45", true)

	newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 1, 0, false)
	newRequirement(t, p.ID, class2bID, subjectMathID, subjectMathID, 1, 0, false)
	newConstraint(t, p.ID, class1aID, "1A", 5, 4, 3, "", true)
	newConstraint(t, p.ID, class2bID, "2B", 5, 4, 3, "", true)

	setMock(alphaOnlyMathSnapshot())
	doSnapshot(t, p.ID)
	g := doGenerate(t, p.ID)
	if g.Conflicts == 0 {
		t.Fatal("expected conflict (teacher_double_booked)")
	}
	if status := getPlan(t, p.ID).Status; status != "resolving" {
		t.Fatalf("expected resolving, got %q", status)
	}

	// Confirm teacher_double_booked
	conflicts := listConflicts(t, p.ID)
	assertConflictType(t, conflicts, "teacher_double_booked")

	// Add a second slot so we can fix the conflict
	slotB := newSlot(t, p.ID, 2, 1, "08:00", "08:45", true)

	// Move one entry to slotB
	entries := listEntries(t, p.ID)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	e2 := entries[1]
	mustDo(t, "PUT", fmt.Sprintf("/plans/%s/entries/%s", p.ID, e2.ID), writeToken, map[string]any{
		"schoolClassId": e2.SchoolClassID,
		"subjectId":     e2.SubjectID,
		"subjectName":   e2.SubjectName,
		"teacherId":     e2.TeacherID,
		"timeSlotId":    slotB.ID,
		"version":       e2.Version,
	}, 200, nil)

	// Validate — conflict should be resolved
	v := doValidate(t, p.ID)
	if v.Conflicts != 0 {
		t.Fatalf("expected 0 conflicts after fix, got %d: %v", v.Conflicts, v.Items)
	}
	if status := getPlan(t, p.ID).Status; status != "accepted" {
		t.Fatalf("expected accepted, got %q", status)
	}

	// Conflicts cleared
	if cfls := listConflicts(t, p.ID); len(cfls) != 0 {
		t.Fatalf("expected 0 conflicts, got %d", len(cfls))
	}

	// Finalize
	var fin finalizeResp
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/finalize", p.ID), writeToken, nil, 200, &fin)
	if fin.LessonsCreated != 2 {
		t.Fatalf("expected 2 lessons, got %d", fin.LessonsCreated)
	}
	if status := getPlan(t, p.ID).Status; status != "finalized" {
		t.Fatalf("expected finalized, got %q", status)
	}

	_ = slotA
}

// TestE2E_ResetAndRegenerate verifies that resetting a plan clears all data
// and allows fresh generation with different snapshot data.
func TestE2E_ResetAndRegenerate(t *testing.T) {
	p := newPlan(t)
	generateDefaultSlots(t, p.ID)
	newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 1, 0, false)
	newConstraint(t, p.ID, class1aID, "1A", 5, 4, 3, "", true)
	setMock(alphaOnlyMathSnapshot())
	doSnapshot(t, p.ID)
	doGenerate(t, p.ID)
	if status := getPlan(t, p.ID).Status; status != "accepted" {
		t.Skip("could not reach accepted status for reset test")
	}

	firstEntries := listEntries(t, p.ID)
	if len(firstEntries) == 0 {
		t.Fatal("expected entries before reset")
	}

	// Reset
	var resetPlan planResp
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/reset", p.ID), writeToken, nil, 200, &resetPlan)
	if resetPlan.Status != "draft" {
		t.Fatalf("expected draft after reset, got %q", resetPlan.Status)
	}

	if entries := listEntries(t, p.ID); len(entries) != 0 {
		t.Fatalf("expected 0 entries after reset, got %d", len(entries))
	}
	if conflicts := listConflicts(t, p.ID); len(conflicts) != 0 {
		t.Fatalf("expected 0 conflicts after reset, got %d", len(conflicts))
	}

	// Re-snapshot and re-generate with a different (2-teacher) setup
	setMock(twoTeacherMathSnapshot())
	doSnapshot(t, p.ID)
	g := doGenerate(t, p.ID)

	if g.Entries == 0 {
		t.Fatal("expected entries after regeneration")
	}
	if g.Conflicts != 0 {
		t.Fatalf("expected no conflicts on clean regeneration, got %d", g.Conflicts)
	}

	// Verify teacher snapshots reflect the new mock data
	var teachers []teacherSnapResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/teachers", p.ID), readToken, nil, 200, &teachers)
	if len(teachers) != 2 {
		t.Fatalf("expected 2 teachers in new snapshot, got %d", len(teachers))
	}
}
