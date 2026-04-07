//go:build integration

package integration_test

import (
	"fmt"
	"testing"
)

func TestPlanCreate_HappyPath(t *testing.T) {
	var p planResp
	mustDo(t, "POST", "/plans", writeToken, map[string]any{
		"schoolYearId": "2025-2026",
		"name":         "Test Plan",
	}, 201, &p)
	t.Cleanup(func() { cleanupPlan(t, p.ID) })

	if p.ID == "" {
		t.Fatal("expected non-empty plan ID")
	}
	if p.Status != "draft" {
		t.Fatalf("expected status=draft, got %q", p.Status)
	}
	if p.Version != 0 {
		t.Fatalf("expected version=0, got %d", p.Version)
	}
	if p.CreatedBy != "test-user-sub" {
		t.Fatalf("expected createdBy=test-user-sub, got %q", p.CreatedBy)
	}
}

func TestPlanCreate_NoToken(t *testing.T) {
	mustDo(t, "POST", "/plans", "", map[string]any{
		"schoolYearId": "2025-2026",
		"name":         "Test",
	}, 401, nil)
}

func TestPlanCreate_ReadTokenForbidden(t *testing.T) {
	mustDo(t, "POST", "/plans", readToken, map[string]any{
		"schoolYearId": "2025-2026",
		"name":         "Test",
	}, 403, nil)
}

func TestPlanCreate_WrongRoleForbidden(t *testing.T) {
	tok := tokenFor("some_other_role")
	mustDo(t, "POST", "/plans", tok, map[string]any{
		"schoolYearId": "2025-2026",
		"name":         "Test",
	}, 403, nil)
}

func TestPlanList(t *testing.T) {
	p1 := newPlan(t)
	p2 := newPlan(t)

	var plans []planResp
	mustDo(t, "GET", "/plans", readToken, nil, 200, &plans)

	found := map[string]bool{}
	for _, p := range plans {
		found[p.ID] = true
	}
	if !found[p1.ID] {
		t.Errorf("plan %s not found in list", p1.ID)
	}
	if !found[p2.ID] {
		t.Errorf("plan %s not found in list", p2.ID)
	}
}

func TestPlanGet_NotFound(t *testing.T) {
	mustDo(t, "GET", "/plans/nonexistent-plan-id", readToken, nil, 404, nil)
}

func TestPlanUpdate_HappyPath(t *testing.T) {
	p := newPlan(t)

	var updated planResp
	mustDo(t, "PUT", fmt.Sprintf("/plans/%s", p.ID), writeToken, map[string]any{
		"schoolYearId": "2025-2026",
		"name":         "Renamed Plan",
		"status":       "draft",
		"version":      p.Version,
	}, 200, &updated)

	if updated.Name != "Renamed Plan" {
		t.Fatalf("expected name=Renamed Plan, got %q", updated.Name)
	}
	if updated.Version != p.Version+1 {
		t.Fatalf("expected version=%d, got %d", p.Version+1, updated.Version)
	}
}

func TestPlanUpdate_OptimisticLock(t *testing.T) {
	p := newPlan(t)

	// First update — must succeed
	var updated planResp
	mustDo(t, "PUT", fmt.Sprintf("/plans/%s", p.ID), writeToken, map[string]any{
		"schoolYearId": "2025-2026",
		"name":         "First",
		"version":      p.Version,
	}, 200, &updated)

	// Second update with same (now stale) version — must fail
	mustDo(t, "PUT", fmt.Sprintf("/plans/%s", p.ID), writeToken, map[string]any{
		"schoolYearId": "2025-2026",
		"name":         "Second",
		"version":      p.Version, // stale!
	}, 409, nil)
}

func TestPlanDelete_Draft(t *testing.T) {
	var p planResp
	mustDo(t, "POST", "/plans", writeToken, map[string]any{
		"schoolYearId": "2025-2026",
		"name":         "To delete",
	}, 201, &p)

	mustDo(t, "DELETE", fmt.Sprintf("/plans/%s", p.ID), writeToken, nil, 204, nil)
	mustDo(t, "GET", fmt.Sprintf("/plans/%s", p.ID), readToken, nil, 404, nil)
}

func TestPlanDelete_NonDraft_Rejected(t *testing.T) {
	// Advance plan to resolving by generating with teacher_double_booked setup
	p := newPlan(t)
	setMock(alphaOnlyMathSnapshot())
	newSlot(t, p.ID, 1, 1, "08:00", "08:45", true) // single slot forces double-booking
	newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 1, 0, false)
	newRequirement(t, p.ID, class2bID, subjectMathID, subjectMathID, 1, 0, false)
	newConstraint(t, p.ID, class1aID, "1A", 5, 4, 3, "", true)
	newConstraint(t, p.ID, class2bID, "2B", 5, 4, 3, "", true)
	doSnapshot(t, p.ID)
	doGenerate(t, p.ID)

	mustDo(t, "DELETE", fmt.Sprintf("/plans/%s", p.ID), writeToken, nil, 400, nil)
}
