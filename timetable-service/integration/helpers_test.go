//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

// ── HTTP primitives ───────────────────────────────────────────────────────────

// do sends a request and returns the response. Caller must close the body.
func do(t *testing.T, method, path, token string, body any) *http.Response {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, srv.URL+path, r)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

// mustDo sends a request, asserts the status code, and optionally decodes the
// JSON response body into out.
func mustDo(t *testing.T, method, path, token string, body any, wantStatus int, out any) {
	t.Helper()
	resp := do(t, method, path, token, body)
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != wantStatus {
		t.Fatalf("%s %s: got status %d, want %d\nbody: %s", method, path, resp.StatusCode, wantStatus, b)
	}
	if out != nil {
		if err := json.Unmarshal(b, out); err != nil {
			t.Fatalf("decode %s %s response: %v\nbody: %s", method, path, err, b)
		}
	}
}

// ── Response types ────────────────────────────────────────────────────────────

type planResp struct {
	ID           string `json:"id"`
	SchoolYearID string `json:"schoolYearId"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	CreatedBy    string `json:"createdBy"`
	Version      int    `json:"version"`
}

type slotResp struct {
	ID        string `json:"id"`
	PlanID    string `json:"planId"`
	DayOfWeek int    `json:"dayOfWeek"`
	Period    int    `json:"period"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	IsMorning bool   `json:"isMorning"`
	Version   int    `json:"version"`
}

type requirementResp struct {
	ID              string `json:"id"`
	PlanID          string `json:"planId"`
	SchoolClassID   string `json:"schoolClassId"`
	SubjectID       string `json:"subjectId"`
	SubjectName     string `json:"subjectName"`
	LessonsPerWeek  int    `json:"lessonsPerWeek"`
	MaxDoubleLessons int   `json:"maxDoubleLessons"`
	PreferMorning   bool   `json:"preferMorning"`
	LessonDurationMin int  `json:"lessonDurationMin"`
	Version         int    `json:"version"`
}

type constraintResp struct {
	ID                string `json:"id"`
	PlanID            string `json:"planId"`
	SchoolClassID     string `json:"schoolClassId"`
	SchoolClassName   string `json:"schoolClassName"`
	MaxEarlyStarts    int    `json:"maxEarlyStarts"`
	MorningPeriods    int    `json:"morningPeriods"`
	AfternoonPeriods  int    `json:"afternoonPeriods"`
	FreeAfternoons    int    `json:"freeAfternoons"`
	FreeAfternoonDays string `json:"freeAfternoonDays"`
	HasTimetable      bool   `json:"hasTimetable"`
	Version           int    `json:"version"`
}

type entryResp struct {
	ID              string `json:"id"`
	PlanID          string `json:"planId"`
	SchoolClassID   string `json:"schoolClassId"`
	SchoolClassName string `json:"schoolClassName"`
	SubjectID       string `json:"subjectId"`
	SubjectName     string `json:"subjectName"`
	TeacherID       string `json:"teacherId"`
	TeacherName     string `json:"teacherName"`
	RoomID          string `json:"roomId"`
	TimeSlotID      string `json:"timeSlotId"`
	IsDoubleLesson  bool   `json:"isDoubleLesson"`
	Version         int    `json:"version"`
}

type conflictResp struct {
	ID            string   `json:"id"`
	PlanID        string   `json:"planId"`
	Type          string   `json:"type"`
	Severity      string   `json:"severity"`
	Description   string   `json:"description"`
	EntryIDs      []string `json:"entryIds"`
	TeacherID     string   `json:"teacherId"`
	SchoolClassID string   `json:"schoolClassId"`
	TimeSlotID    string   `json:"timeSlotId"`
	Resolved      bool     `json:"resolved"`
	Version       int      `json:"version"`
}

type teacherSnapResp struct {
	ID       string   `json:"id"`
	PlanID   string   `json:"planId"`
	Name     string   `json:"name"`
	Prename  string   `json:"prename"`
	Subjects []string `json:"subjects"`
}

type snapshotSummaryResp struct {
	Teachers int `json:"teachers"`
	Subjects int `json:"subjects"`
	Classes  int `json:"classes"`
	Rooms    int `json:"rooms"`
}

type generateResp struct {
	Entries   int `json:"entries"`
	Conflicts int `json:"conflicts"`
}

type validateResp struct {
	Conflicts int            `json:"conflicts"`
	Items     []conflictResp `json:"items"`
}

type finalizeResp struct {
	LessonsCreated int `json:"lessonsCreated"`
}

// ── Resource creation helpers ─────────────────────────────────────────────────

// newPlan creates a draft plan and registers cleanup. Returns the plan.
func newPlan(t *testing.T) planResp {
	t.Helper()
	var p planResp
	mustDo(t, "POST", "/plans", writeToken, map[string]any{
		"schoolYearId": "2025-2026",
		"name":         t.Name(),
	}, 201, &p)
	t.Cleanup(func() { cleanupPlan(t, p.ID) })
	return p
}

// cleanupPlan resets then deletes a plan, ignoring errors.
func cleanupPlan(t *testing.T, planID string) {
	t.Helper()
	resp := do(t, "POST", "/plans/"+planID+"/reset", writeToken, nil)
	resp.Body.Close()
	resp = do(t, "DELETE", "/plans/"+planID, writeToken, nil)
	resp.Body.Close()
}

// newSlot creates a single time slot for the given plan.
func newSlot(t *testing.T, planID string, day, period int, start, end string, morning bool) slotResp {
	t.Helper()
	var s slotResp
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/time-slots", planID), writeToken, map[string]any{
		"dayOfWeek": day,
		"period":    period,
		"startTime": start,
		"endTime":   end,
		"isMorning": morning,
	}, 201, &s)
	return s
}

// generateDefaultSlots generates the default Mon-Fri slot grid.
func generateDefaultSlots(t *testing.T, planID string) []slotResp {
	t.Helper()
	var slots []slotResp
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/time-slots/generate-default", planID), writeToken, map[string]any{
		"morningPeriods":    4,
		"afternoonPeriods":  3,
		"startTime":         "07:45",
		"lessonDurationMin": 45,
		"breakDurationMin":  5,
		"lunchBreakMin":     60,
	}, 201, &slots)
	return slots
}

// newRequirement creates a grade requirement.
func newRequirement(t *testing.T, planID, classID, subjectID, subjectName string, lpw, maxDouble int, preferMorning bool) requirementResp {
	t.Helper()
	var r requirementResp
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/requirements", planID), writeToken, map[string]any{
		"schoolClassId":    classID,
		"subjectId":        subjectID,
		"subjectName":      subjectName,
		"lessonsPerWeek":   lpw,
		"maxDoubleLessons": maxDouble,
		"preferMorning":    preferMorning,
		"lessonDurationMin": 45,
	}, 201, &r)
	return r
}

// newConstraint creates a class constraint.
func newConstraint(t *testing.T, planID, classID, className string, maxEarly, morningP, afternoonP int, freeDays string, hasTimetable bool) constraintResp {
	t.Helper()
	var c constraintResp
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/constraints", planID), writeToken, map[string]any{
		"schoolClassId":     classID,
		"schoolClassName":   className,
		"maxEarlyStarts":    maxEarly,
		"morningPeriods":    morningP,
		"afternoonPeriods":  afternoonP,
		"freeAfternoonDays": freeDays,
		"hasTimetable":      hasTimetable,
	}, 201, &c)
	return c
}

// doSnapshot calls /snapshot and returns the summary.
func doSnapshot(t *testing.T, planID string) snapshotSummaryResp {
	t.Helper()
	var s snapshotSummaryResp
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/snapshot", planID), writeToken, nil, 200, &s)
	return s
}

// doGenerate calls /generate and returns the result.
func doGenerate(t *testing.T, planID string) generateResp {
	t.Helper()
	var g generateResp
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/generate", planID), writeToken, nil, 200, &g)
	return g
}

// doValidate calls /validate and returns the result.
func doValidate(t *testing.T, planID string) validateResp {
	t.Helper()
	var v validateResp
	mustDo(t, "POST", fmt.Sprintf("/plans/%s/validate", planID), writeToken, nil, 200, &v)
	return v
}

// getPlan returns the current state of the plan.
func getPlan(t *testing.T, planID string) planResp {
	t.Helper()
	var p planResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s", planID), readToken, nil, 200, &p)
	return p
}

// listEntries returns all entries for a plan.
func listEntries(t *testing.T, planID string) []entryResp {
	t.Helper()
	var entries []entryResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/entries", planID), readToken, nil, 200, &entries)
	return entries
}

// listConflicts returns all conflicts for a plan.
func listConflicts(t *testing.T, planID string) []conflictResp {
	t.Helper()
	var conflicts []conflictResp
	mustDo(t, "GET", fmt.Sprintf("/plans/%s/conflicts", planID), readToken, nil, 200, &conflicts)
	return conflicts
}

// assertConflictType asserts that the list contains a conflict of the given type
// and returns it.
func assertConflictType(t *testing.T, conflicts []conflictResp, wantType string) conflictResp {
	t.Helper()
	for _, c := range conflicts {
		if c.Type == wantType {
			return c
		}
	}
	types := make([]string, len(conflicts))
	for i, c := range conflicts {
		types[i] = c.Type
	}
	t.Fatalf("no conflict of type %q found; got types: %v", wantType, types)
	return conflictResp{}
}

// assertNoConflictType asserts that the list does NOT contain a conflict of the
// given type.
func assertNoConflictType(t *testing.T, conflicts []conflictResp, wantType string) {
	t.Helper()
	for _, c := range conflicts {
		if c.Type == wantType {
			t.Fatalf("unexpected conflict of type %q found", wantType)
		}
	}
}
