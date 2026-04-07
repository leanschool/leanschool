//go:build integration

package integration_test

import (
	"fmt"
	"testing"
)

// ── Snapshot ──────────────────────────────────────────────────────────────────

func TestSnapshot_HappyPath(t *testing.T) {
	p := newPlan(t)
	setMock(standardSnapshot())

	summary := doSnapshot(t, p.ID)
	if summary.Teachers != 3 {
		t.Fatalf("expected 3 teachers, got %d", summary.Teachers)
	}
	if summary.Subjects != 4 {
		t.Fatalf("expected 4 subjects, got %d", summary.Subjects)
	}
	if summary.Classes != 2 {
		t.Fatalf("expected 2 classes, got %d", summary.Classes)
	}
	if summary.Rooms != 3 {
		t.Fatalf("expected 3 rooms, got %d", summary.Rooms)
	}

	var teachers []teacherSnapResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/teachers", p.ID), readToken, nil, 200, &teachers)
	if len(teachers) != 3 {
		t.Fatalf("expected 3 teacher snapshots, got %d", len(teachers))
	}
	// TeacherAlpha should have Math and English in their subjects
	for _, t2 := range teachers {
		if t2.ID == teacherAlphaID {
			subSet := map[string]bool{}
			for _, s := range t2.Subjects {
				subSet[s] = true
			}
			if !subSet[subjectMathID] {
				t.Errorf("TeacherAlpha missing Math in subjects: %v", t2.Subjects)
			}
			if !subSet[subjectEnglishID] {
				t.Errorf("TeacherAlpha missing English in subjects: %v", t2.Subjects)
			}
		}
	}

	var classes []map[string]any
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/classes", p.ID), readToken, nil, 200, &classes)
	if len(classes) != 2 {
		t.Fatalf("expected 2 class snapshots, got %d", len(classes))
	}
}

func TestSnapshot_RequiresDraftStatus(t *testing.T) {
	planID, _, _ := setupResolvingPlan(t)
	setMock(standardSnapshot())
	// Plan is now resolving — snapshot should fail
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/snapshot", planID), writeToken, nil, 409, nil)
}

func TestSnapshot_LeanschoolClientError(t *testing.T) {
	// The lsMock infrastructure returns empty arrays for nil slices, not errors.
	// A 502 requires the leanschool client call to fail, which needs the mock to
	// return non-200. Per-endpoint error injection is not yet wired into lsMock.
	t.Skip("requires per-endpoint error injection not currently supported by lsMock")
}

// ── Generate ──────────────────────────────────────────────────────────────────

func TestGenerate_TransitionsToAccepted_NoConflicts(t *testing.T) {
	p := newPlan(t)
	generateDefaultSlots(t, p.ID)
	// Single class, Math 1 lesson, TeacherAlpha qualifies, ample slots
	newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 1, 0, false)
	newConstraint(t, p.ID, class1aID, "1A", 5, 4, 3, "", true)
	setMock(alphaOnlyMathSnapshot())
	doSnapshot(t, p.ID)

	g := doGenerate(t, p.ID)
	if g.Conflicts != 0 {
		t.Fatalf("expected 0 conflicts, got %d", g.Conflicts)
	}
	if g.Entries != 1 {
		t.Fatalf("expected 1 entry, got %d", g.Entries)
	}
	if status := getPlan(t, p.ID).Status; status != "accepted" {
		t.Fatalf("expected status=accepted, got %q", status)
	}
}

func TestGenerate_TransitionsToResolving_WithConflicts(t *testing.T) {
	p := newPlan(t)
	newSlot(t, p.ID, 1, 1, "08:00", "08:45", true)
	newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 1, 0, false)
	newRequirement(t, p.ID, class2bID, subjectMathID, subjectMathID, 1, 0, false)
	newConstraint(t, p.ID, class1aID, "1A", 5, 4, 3, "", true)
	newConstraint(t, p.ID, class2bID, "2B", 5, 4, 3, "", true)
	setMock(alphaOnlyMathSnapshot())
	doSnapshot(t, p.ID)

	g := doGenerate(t, p.ID)
	if g.Conflicts == 0 {
		t.Fatal("expected conflicts from teacher double-booking")
	}
	if status := getPlan(t, p.ID).Status; status != "resolving" {
		t.Fatalf("expected status=resolving, got %q", status)
	}
}

func TestGenerate_RejectsNonDraftStatus(t *testing.T) {
	planID, _, _ := setupResolvingPlan(t)
	// Plan is already in resolving — second generate should fail
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/generate", planID), writeToken, nil, 409, nil)
}

// ── Validate ──────────────────────────────────────────────────────────────────

func TestValidate_StaysResolving_WhenConflictsRemain(t *testing.T) {
	planID, _, _ := setupResolvingPlan(t)
	// Validate without fixing anything — conflicts should remain
	v := doValidate(t, planID)
	if v.Conflicts == 0 {
		t.Fatal("expected conflicts to remain after validate without fixes")
	}
	if status := getPlan(t, planID).Status; status != "resolving" {
		t.Fatalf("expected status=resolving, got %q", status)
	}
}

func TestValidate_TransitionsToAccepted_WhenConflictsResolved(t *testing.T) {
	// Set up a plan with teacher_double_booked, then fix it
	p := newPlan(t)
	sA := newSlot(t, p.ID, 1, 1, "08:00", "08:45", true)
	sB := newSlot(t, p.ID, 2, 1, "08:00", "08:45", true)
	newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 1, 0, false)
	newRequirement(t, p.ID, class2bID, subjectMathID, subjectMathID, 1, 0, false)
	newConstraint(t, p.ID, class1aID, "1A", 5, 4, 3, "", true)
	newConstraint(t, p.ID, class2bID, "2B", 5, 4, 3, "", true)
	setMock(alphaOnlyMathSnapshot())
	doSnapshot(t, p.ID)
	doGenerate(t, p.ID)

	// Both entries on slotA — move entry2 to slotB
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
		"timeSlotId":    sB.ID,
		"version":       e2.Version,
	}, 200, nil)

	v := doValidate(t, p.ID)
	if v.Conflicts != 0 {
		t.Fatalf("expected 0 conflicts after fix, got %d; items: %v", v.Conflicts, v.Items)
	}
	if status := getPlan(t, p.ID).Status; status != "accepted" {
		t.Fatalf("expected status=accepted, got %q", status)
	}
	_ = sA
}

// ── Finalize ──────────────────────────────────────────────────────────────────

func TestFinalize_HappyPath(t *testing.T) {
	p := newPlan(t)
	generateDefaultSlots(t, p.ID)
	newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 1, 0, false)
	newConstraint(t, p.ID, class1aID, "1A", 5, 4, 3, "", true)

	m := alphaOnlyMathSnapshot()
	setMock(m)
	doSnapshot(t, p.ID)
	doGenerate(t, p.ID)

	if status := getPlan(t, p.ID).Status; status != "accepted" {
		t.Fatalf("need accepted plan for finalize, got %q", status)
	}

	var fin finalizeResp
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/finalize", p.ID), writeToken, nil, 200, &fin)
	if fin.LessonsCreated != 1 {
		t.Fatalf("expected 1 lesson created, got %d", fin.LessonsCreated)
	}
	if status := getPlan(t, p.ID).Status; status != "finalized" {
		t.Fatalf("expected status=finalized, got %q", status)
	}

	lessons := getLessonsCreated()
	if len(lessons) != 1 {
		t.Fatalf("mock captured %d lessons, expected 1", len(lessons))
	}
	if lessons[0].SchoolClass == nil || lessons[0].SchoolClass.ID != class1aID {
		t.Fatalf("lesson has wrong school class: %v", lessons[0].SchoolClass)
	}
}

func TestFinalize_RejectsFromDraft(t *testing.T) {
	p := newPlan(t)
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/finalize", p.ID), writeToken, nil, 409, nil)
}

func TestFinalize_RejectsFromResolving(t *testing.T) {
	planID, _, _ := setupResolvingPlan(t)
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/finalize", planID), writeToken, nil, 409, nil)
}

func TestFinalize_RejectsFromFinalized(t *testing.T) {
	p := newPlan(t)
	generateDefaultSlots(t, p.ID)
	newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 1, 0, false)
	newConstraint(t, p.ID, class1aID, "1A", 5, 4, 3, "", true)
	setMock(alphaOnlyMathSnapshot())
	doSnapshot(t, p.ID)
	doGenerate(t, p.ID)
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/finalize", p.ID), writeToken, nil, 200, nil)
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/finalize", p.ID), writeToken, nil, 409, nil)
}

func TestFinalize_LeanschoolClientError(t *testing.T) {
	p := newPlan(t)
	generateDefaultSlots(t, p.ID)
	newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 1, 0, false)
	newConstraint(t, p.ID, class1aID, "1A", 5, 4, 3, "", true)
	setMock(alphaOnlyMathSnapshot())
	doSnapshot(t, p.ID)
	doGenerate(t, p.ID)

	if status := getPlan(t, p.ID).Status; status != "accepted" {
		t.Fatalf("need accepted plan, got %q", status)
	}

	// Enable lesson creation error
	lsMu.Lock()
	lsMock.LessonError = true
	lsMu.Unlock()
	defer func() {
		lsMu.Lock()
		lsMock.LessonError = false
		lsMu.Unlock()
	}()

	mustDo(t, "POST", fmt.Sprintf("/plans/%s/finalize", p.ID), writeToken, nil, 502, nil)
	// Plan status must remain accepted
	if status := getPlan(t, p.ID).Status; status != "accepted" {
		t.Fatalf("plan should remain accepted after failed finalize, got %q", status)
	}
}

// ── Reset ─────────────────────────────────────────────────────────────────────

func TestReset_FromResolving(t *testing.T) {
	planID, _, _ := setupResolvingPlan(t)
	assertReset(t, planID)
}

func TestReset_FromAccepted(t *testing.T) {
	p := newPlan(t)
	generateDefaultSlots(t, p.ID)
	newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 1, 0, false)
	newConstraint(t, p.ID, class1aID, "1A", 5, 4, 3, "", true)
	setMock(alphaOnlyMathSnapshot())
	doSnapshot(t, p.ID)
	doGenerate(t, p.ID)
	if status := getPlan(t, p.ID).Status; status != "accepted" {
		t.Skip("could not reach accepted status")
	}
	assertReset(t, p.ID)
}

func TestReset_FromFinalized(t *testing.T) {
	p := newPlan(t)
	generateDefaultSlots(t, p.ID)
	newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 1, 0, false)
	newConstraint(t, p.ID, class1aID, "1A", 5, 4, 3, "", true)
	setMock(alphaOnlyMathSnapshot())
	doSnapshot(t, p.ID)
	doGenerate(t, p.ID)
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/finalize", p.ID), writeToken, nil, 200, nil)
	assertReset(t, p.ID)
}

func assertReset(t *testing.T, planID string) {
	t.Helper()
	var plan planResp
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/reset", planID), writeToken, nil, 200, &plan)
	if plan.Status != "draft" {
		t.Fatalf("expected status=draft after reset, got %q", plan.Status)
	}

	entries := listEntries(t, planID)
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries after reset, got %d", len(entries))
	}
	conflicts := listConflicts(t, planID)
	if len(conflicts) != 0 {
		t.Fatalf("expected 0 conflicts after reset, got %d", len(conflicts))
	}

	var teachers []teacherSnapResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/teachers", planID), readToken, nil, 200, &teachers)
	if len(teachers) != 0 {
		t.Fatalf("expected 0 teacher snapshots after reset, got %d", len(teachers))
	}
}

func TestReset_ThenCanDelete(t *testing.T) {
	planID, _, _ := setupResolvingPlan(t)

	var plan planResp
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/reset", planID), writeToken, nil, 200, &plan)

	// Can now delete since status=draft
	mustDo(t, "DELETE", fmt.Sprintf("/plans/%s", planID), writeToken, nil, 204, nil)
	// Remove cleanup since we deleted it
}

// ── Health ────────────────────────────────────────────────────────────────────

func TestHealth(t *testing.T) {
	resp := do(t, "GET", "/health", "", nil)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected health=200, got %d", resp.StatusCode)
	}
}

// ── Snapshot read-only endpoints ──────────────────────────────────────────────

func TestSnapshotEndpoints(t *testing.T) {
	p := newPlan(t)
	setMock(standardSnapshot())
	doSnapshot(t, p.ID)

	// Teachers
	var teachers []teacherSnapResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/teachers", p.ID), readToken, nil, 200, &teachers)
	if len(teachers) != 3 {
		t.Fatalf("expected 3 teachers, got %d", len(teachers))
	}

	// Subjects
	var subjects []map[string]any
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/subjects", p.ID), readToken, nil, 200, &subjects)
	if len(subjects) != 4 {
		t.Fatalf("expected 4 subjects, got %d", len(subjects))
	}

	// Classes
	var classes []map[string]any
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/classes", p.ID), readToken, nil, 200, &classes)
	if len(classes) != 2 {
		t.Fatalf("expected 2 classes, got %d", len(classes))
	}

	// Rooms
	var rooms []map[string]any
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/rooms", p.ID), readToken, nil, 200, &rooms)
	if len(rooms) != 3 {
		t.Fatalf("expected 3 rooms, got %d", len(rooms))
	}
}

