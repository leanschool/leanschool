//go:build integration

package integration_test

import (
	"fmt"
	"testing"
)

func TestConstraint_CRUD(t *testing.T) {
	p := newPlan(t)

	// Create
	c := newConstraint(t, p.ID, class1aID, "1A", 2, 4, 3, "5", true)
	if c.ID == "" {
		t.Fatal("expected non-empty constraint ID")
	}
	if c.MaxEarlyStarts != 2 {
		t.Fatalf("expected maxEarlyStarts=2, got %d", c.MaxEarlyStarts)
	}
	if c.FreeAfternoonDays != "5" {
		t.Fatalf("expected freeAfternoonDays=5, got %q", c.FreeAfternoonDays)
	}

	// List
	var cons []constraintResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/constraints", p.ID), readToken, nil, 200, &cons)
	found := false
	for _, cc := range cons {
		if cc.ID == c.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("created constraint not found in list")
	}

	// Update
	var updated constraintResp
	mustDo(t, "PUT", fmt.Sprintf("/plans/%s/constraints/%s", p.ID, c.ID), writeToken, map[string]any{
		"schoolClassId":     class1aID,
		"schoolClassName":   "1A",
		"maxEarlyStarts":    3,
		"morningPeriods":    4,
		"afternoonPeriods":  3,
		"freeAfternoonDays": "",
		"hasTimetable":      true,
		"version":           c.Version,
	}, 200, &updated)
	if updated.MaxEarlyStarts != 3 {
		t.Fatalf("expected maxEarlyStarts=3, got %d", updated.MaxEarlyStarts)
	}
	if updated.Version != c.Version+1 {
		t.Fatalf("expected version=%d, got %d", c.Version+1, updated.Version)
	}

	// Delete
	mustDo(t, "DELETE", fmt.Sprintf("/plans/%s/constraints/%s", p.ID, c.ID), writeToken, nil, 204, nil)

	var afterDelete []constraintResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/constraints", p.ID), readToken, nil, 200, &afterDelete)
	for _, cc := range afterDelete {
		if cc.ID == c.ID {
			t.Fatal("deleted constraint still in list")
		}
	}
}

func TestConstraint_OptimisticLock(t *testing.T) {
	p := newPlan(t)
	c := newConstraint(t, p.ID, class1aID, "1A", 2, 4, 3, "", true)

	body := map[string]any{
		"schoolClassId":     class1aID,
		"schoolClassName":   "1A",
		"maxEarlyStarts":    3,
		"morningPeriods":    4,
		"afternoonPeriods":  3,
		"freeAfternoonDays": "",
		"hasTimetable":      true,
		"version":           c.Version,
	}
	mustDo(t, "PUT", fmt.Sprintf("/plans/%s/constraints/%s", p.ID, c.ID), writeToken, body, 200, nil)
	mustDo(t, "PUT", fmt.Sprintf("/plans/%s/constraints/%s", p.ID, c.ID), writeToken, body, 409, nil) // stale
}

// TestConstraint_DuplicateClass tests that creating two constraints for the same
// class+plan returns an error (UNIQUE constraint in DB).
func TestConstraint_DuplicateClass(t *testing.T) {
	p := newPlan(t)
	newConstraint(t, p.ID, class1aID, "1A", 2, 4, 3, "", true)

	resp := do(t, "POST", fmt.Sprintf("/plans/%s/constraints", p.ID), writeToken, map[string]any{
		"schoolClassId":     class1aID, // same class again!
		"schoolClassName":   "1A",
		"maxEarlyStarts":    1,
		"morningPeriods":    4,
		"afternoonPeriods":  3,
		"freeAfternoonDays": "",
		"hasTimetable":      true,
	})
	resp.Body.Close()
	if resp.StatusCode == 201 {
		t.Fatal("expected error for duplicate class constraint, got 201")
	}
}
