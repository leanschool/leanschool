//go:build integration

package integration_test

import (
	"fmt"
	"testing"
)

func TestTimeSlot_CRUD(t *testing.T) {
	p := newPlan(t)

	// Create
	var s slotResp
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/time-slots", p.ID), writeToken, map[string]any{
		"dayOfWeek": 1,
		"period":    1,
		"startTime": "08:00",
		"endTime":   "08:45",
		"isMorning": true,
	}, 201, &s)
	if s.ID == "" {
		t.Fatal("expected non-empty slot ID")
	}
	if s.PlanID != p.ID {
		t.Fatalf("planId mismatch: got %q, want %q", s.PlanID, p.ID)
	}

	// List
	var slots []slotResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/time-slots", p.ID), readToken, nil, 200, &slots)
	found := false
	for _, sl := range slots {
		if sl.ID == s.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("created slot not found in list")
	}

	// Update
	var updated slotResp
	mustDo(t, "PUT", fmt.Sprintf("/plans/%s/time-slots/%s", p.ID, s.ID), writeToken, map[string]any{
		"dayOfWeek": 1,
		"period":    1,
		"startTime": "08:15",
		"endTime":   "09:00",
		"isMorning": true,
		"version":   s.Version,
	}, 200, &updated)
	if updated.StartTime != "08:15" {
		t.Fatalf("expected startTime=08:15, got %q", updated.StartTime)
	}
	if updated.Version != s.Version+1 {
		t.Fatalf("expected version=%d, got %d", s.Version+1, updated.Version)
	}

	// Delete
	mustDo(t, "DELETE", fmt.Sprintf("/plans/%s/time-slots/%s", p.ID, s.ID), writeToken, nil, 204, nil)

	// Verify gone
	var afterDelete []slotResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/time-slots", p.ID), readToken, nil, 200, &afterDelete)
	for _, sl := range afterDelete {
		if sl.ID == s.ID {
			t.Fatal("deleted slot still appears in list")
		}
	}
}

func TestTimeSlot_GenerateDefault(t *testing.T) {
	p := newPlan(t)
	slots := generateDefaultSlots(t, p.ID)

	// 5 days × 7 periods = 35 slots
	if len(slots) != 35 {
		t.Fatalf("expected 35 slots, got %d", len(slots))
	}

	// Find period 1 on Monday (dayOfWeek=1)
	var mon1, mon2 *slotResp
	for i := range slots {
		if slots[i].DayOfWeek == 1 && slots[i].Period == 1 {
			mon1 = &slots[i]
		}
		if slots[i].DayOfWeek == 1 && slots[i].Period == 2 {
			mon2 = &slots[i]
		}
	}
	if mon1 == nil {
		t.Fatal("Monday period 1 slot not found")
	}
	if mon1.StartTime != "07:45" {
		t.Fatalf("expected period 1 start=07:45, got %q", mon1.StartTime)
	}
	if mon1.EndTime != "08:30" {
		t.Fatalf("expected period 1 end=08:30, got %q", mon1.EndTime)
	}
	if !mon1.IsMorning {
		t.Fatal("period 1 should be morning")
	}
	if mon2 == nil {
		t.Fatal("Monday period 2 slot not found")
	}
	if mon2.StartTime != "08:35" {
		t.Fatalf("expected period 2 start=08:35, got %q", mon2.StartTime)
	}

	// Periods 5-7 should be afternoon
	for _, sl := range slots {
		if sl.DayOfWeek == 1 {
			if sl.Period <= 4 && !sl.IsMorning {
				t.Errorf("period %d should be morning", sl.Period)
			}
			if sl.Period > 4 && sl.IsMorning {
				t.Errorf("period %d should be afternoon", sl.Period)
			}
		}
	}
}

func TestTimeSlot_OptimisticLock(t *testing.T) {
	p := newPlan(t)
	s := newSlot(t, p.ID, 2, 1, "08:00", "08:45", true)

	// First update succeeds
	mustDo(t, "PUT", fmt.Sprintf("/plans/%s/time-slots/%s", p.ID, s.ID), writeToken, map[string]any{
		"dayOfWeek": 2, "period": 1, "startTime": "08:00", "endTime": "08:45",
		"isMorning": true, "version": s.Version,
	}, 200, nil)

	// Second update with stale version fails
	mustDo(t, "PUT", fmt.Sprintf("/plans/%s/time-slots/%s", p.ID, s.ID), writeToken, map[string]any{
		"dayOfWeek": 2, "period": 1, "startTime": "08:00", "endTime": "08:45",
		"isMorning": true, "version": s.Version, // stale
	}, 409, nil)
}
