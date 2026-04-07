//go:build integration

package integration_test

import (
	"fmt"
	"testing"
)

func TestConflictList_Empty(t *testing.T) {
	p := newPlan(t)
	var conflicts []conflictResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/conflicts", p.ID), readToken, nil, 200, &conflicts)
	if len(conflicts) != 0 {
		t.Fatalf("expected empty conflict list, got %d", len(conflicts))
	}
}

func TestConflictList_AfterGenerate(t *testing.T) {
	planID, _, _ := setupResolvingPlan(t)
	conflicts := listConflicts(t, planID)
	if len(conflicts) == 0 {
		t.Fatal("expected conflicts after generate with insufficient slots")
	}
	for _, c := range conflicts {
		if c.Severity != "error" {
			t.Errorf("expected severity=error, got %q", c.Severity)
		}
		if c.ID == "" {
			t.Error("conflict has empty ID")
		}
	}
}

func TestConflictList_FilterResolved_False(t *testing.T) {
	planID, _, _ := setupResolvingPlan(t)
	conflicts := listConflicts(t, planID)
	if len(conflicts) == 0 {
		t.Skip("no conflicts to filter")
	}

	var filtered []conflictResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/conflicts?resolved=false", planID), readToken, nil, 200, &filtered)
	// All generated conflicts are unresolved
	if len(filtered) != len(conflicts) {
		t.Fatalf("?resolved=false returned %d, expected %d", len(filtered), len(conflicts))
	}
}

func TestConflictList_FilterResolved_True(t *testing.T) {
	planID, _, _ := setupResolvingPlan(t)

	// No conflicts have been manually resolved → expect empty result
	var filtered []conflictResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/conflicts?resolved=true", planID), readToken, nil, 200, &filtered)
	if len(filtered) != 0 {
		t.Fatalf("?resolved=true returned %d items, expected 0 (nothing manually resolved)", len(filtered))
	}
}

func TestConflictList_FilterByTeacher(t *testing.T) {
	planID, _, _ := setupDoubleBookedPlan(t)
	conflicts := listConflicts(t, planID)
	if len(conflicts) == 0 {
		t.Skip("no conflicts")
	}

	var filtered []conflictResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/conflicts?teacherId=%s", planID, teacherAlphaID), readToken, nil, 200, &filtered)
	for _, c := range filtered {
		if c.TeacherID != teacherAlphaID {
			t.Fatalf("filter returned conflict for teacher %q, expected %s", c.TeacherID, teacherAlphaID)
		}
	}
	if len(filtered) == 0 {
		t.Fatal("expected at least 1 conflict for teacherAlpha")
	}
}

func TestConflictList_ForbiddenWithoutToken(t *testing.T) {
	p := newPlan(t)
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/conflicts", p.ID), "", nil, 401, nil)
}
