//go:build integration

package integration_test

import (
	"fmt"
	"testing"
)

func TestRequirement_CRUD(t *testing.T) {
	p := newPlan(t)

	// Create
	r := newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 3, 1, true)
	if r.ID == "" {
		t.Fatal("expected non-empty requirement ID")
	}
	if r.LessonsPerWeek != 3 {
		t.Fatalf("expected lessonsPerWeek=3, got %d", r.LessonsPerWeek)
	}

	// List
	var reqs []requirementResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/requirements", p.ID), readToken, nil, 200, &reqs)
	found := false
	for _, rr := range reqs {
		if rr.ID == r.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("created requirement not found in list")
	}

	// Update
	var updated requirementResp
	mustDo(t, "PUT", fmt.Sprintf("/plans/%s/requirements/%s", p.ID, r.ID), writeToken, map[string]any{
		"schoolClassId":    class1aID,
		"subjectId":        subjectMathID,
		"subjectName":      subjectMathID,
		"lessonsPerWeek":   2,
		"maxDoubleLessons": 0,
		"preferMorning":    false,
		"lessonDurationMin": 45,
		"version":          r.Version,
	}, 200, &updated)
	if updated.LessonsPerWeek != 2 {
		t.Fatalf("expected lessonsPerWeek=2, got %d", updated.LessonsPerWeek)
	}
	if updated.Version != r.Version+1 {
		t.Fatalf("expected version=%d, got %d", r.Version+1, updated.Version)
	}

	// Delete
	mustDo(t, "DELETE", fmt.Sprintf("/plans/%s/requirements/%s", p.ID, r.ID), writeToken, nil, 204, nil)

	var afterDelete []requirementResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/requirements", p.ID), readToken, nil, 200, &afterDelete)
	for _, rr := range afterDelete {
		if rr.ID == r.ID {
			t.Fatal("deleted requirement still in list")
		}
	}
}

func TestRequirement_OptimisticLock(t *testing.T) {
	p := newPlan(t)
	r := newRequirement(t, p.ID, class1aID, subjectMathID, subjectMathID, 2, 0, false)

	updateBody := map[string]any{
		"schoolClassId":    class1aID,
		"subjectId":        subjectMathID,
		"subjectName":      subjectMathID,
		"lessonsPerWeek":   3,
		"maxDoubleLessons": 0,
		"lessonDurationMin": 45,
		"version":          r.Version,
	}
	mustDo(t, "PUT", fmt.Sprintf("/plans/%s/requirements/%s", p.ID, r.ID), writeToken, updateBody, 200, nil)
	mustDo(t, "PUT", fmt.Sprintf("/plans/%s/requirements/%s", p.ID, r.ID), writeToken, updateBody, 409, nil) // stale
}
