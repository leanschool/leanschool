package planner

import (
	"fmt"

	"github.com/Joel-Haeberli/timetable-service/internal/model"
)

// Validate detects scheduling conflicts in the given timetable entries.
// It returns a slice of Conflict structs (without PlanID set — the caller
// must set PlanID before persisting).
func Validate(
	entries []*model.TimetableEntry,
	requirements []*model.GradeRequirement,
	constraints []*model.ClassConstraint,
	teachers []*model.TeacherSnapshot,
	timeSlots []*model.TimeSlotDefinition,
) []*model.Conflict {
	var conflicts []*model.Conflict
	conflictCounter := 0

	newConflict := func(ct model.ConflictType, desc string, entryIDs []string) *model.Conflict {
		conflictCounter++
		return &model.Conflict{
			ID:       fmt.Sprintf("conflict-%d", conflictCounter),
			Type:     ct,
			Severity: model.SeverityError,
			Description: desc,
			EntryIDs: entryIDs,
		}
	}

	// Build lookups
	slotByID := make(map[string]*model.TimeSlotDefinition, len(timeSlots))
	for _, ts := range timeSlots {
		slotByID[ts.ID] = ts
	}

	teacherByID := make(map[string]*model.TeacherSnapshot, len(teachers))
	teacherSubjects := make(map[string]map[string]bool) // teacherID → set of subjectIDs
	for _, t := range teachers {
		teacherByID[t.ID] = t
		subs := make(map[string]bool, len(t.Subjects))
		for _, sid := range t.Subjects {
			subs[sid] = true
		}
		teacherSubjects[t.ID] = subs
	}

	conByClass := make(map[string]*model.ClassConstraint, len(constraints))
	for _, c := range constraints {
		conByClass[c.SchoolClassID] = c
	}

	reqLookup := make(map[string]*model.GradeRequirement) // classID|subjectID → requirement
	for _, r := range requirements {
		key := r.SchoolClassID + "|" + r.SubjectID
		reqLookup[key] = r
	}

	// ── 1. Teacher double-booked ────────────────────────────────────────
	type teacherSlotKey struct {
		teacherID string
		slotID    string
	}
	teacherSlotEntries := make(map[teacherSlotKey][]string)
	for _, e := range entries {
		if e.TeacherID == "" {
			continue
		}
		key := teacherSlotKey{e.TeacherID, e.TimeSlotID}
		teacherSlotEntries[key] = append(teacherSlotEntries[key], e.ID)
	}
	for key, eids := range teacherSlotEntries {
		if len(eids) > 1 {
			t := teacherByID[key.teacherID]
			tName := key.teacherID
			if t != nil {
				tName = t.Prename + " " + t.Name
			}
			ts := slotByID[key.slotID]
			slotDesc := key.slotID
			if ts != nil {
				slotDesc = fmt.Sprintf("day %d period %d", ts.DayOfWeek, ts.Period)
			}
			c := newConflict(
				model.ConflictTeacherDoubleBooked,
				fmt.Sprintf("Teacher %s is double-booked at %s", tName, slotDesc),
				eids,
			)
			c.TeacherID = key.teacherID
			c.TimeSlotID = key.slotID
			conflicts = append(conflicts, c)
		}
	}

	// ── 2. Room double-booked ───────────────────────────────────────────
	type roomSlotKey struct {
		roomID string
		slotID string
	}
	roomSlotEntries := make(map[roomSlotKey][]string)
	for _, e := range entries {
		if e.RoomID == "" {
			continue
		}
		key := roomSlotKey{e.RoomID, e.TimeSlotID}
		roomSlotEntries[key] = append(roomSlotEntries[key], e.ID)
	}
	for key, eids := range roomSlotEntries {
		if len(eids) > 1 {
			ts := slotByID[key.slotID]
			slotDesc := key.slotID
			if ts != nil {
				slotDesc = fmt.Sprintf("day %d period %d", ts.DayOfWeek, ts.Period)
			}
			c := newConflict(
				model.ConflictRoomDoubleBooked,
				fmt.Sprintf("Room %s is double-booked at %s", key.roomID, slotDesc),
				eids,
			)
			c.TimeSlotID = key.slotID
			conflicts = append(conflicts, c)
		}
	}

	// ── 3. Teacher not qualified ────────────────────────────────────────
	for _, e := range entries {
		if e.TeacherID == "" {
			continue
		}
		subs, ok := teacherSubjects[e.TeacherID]
		if !ok || !subs[e.SubjectID] {
			t := teacherByID[e.TeacherID]
			tName := e.TeacherID
			if t != nil {
				tName = t.Prename + " " + t.Name
			}
			c := newConflict(
				model.ConflictTeacherNotQualified,
				fmt.Sprintf("Teacher %s is not qualified to teach %s", tName, e.SubjectName),
				[]string{e.ID},
			)
			c.TeacherID = e.TeacherID
			c.SchoolClassID = e.SchoolClassID
			conflicts = append(conflicts, c)
		}
	}

	// ── 4. Too many/few lessons ─────────────────────────────────────────
	type classSubjectKey struct {
		classID   string
		subjectID string
	}
	lessonCounts := make(map[classSubjectKey][]string) // key → entry IDs
	for _, e := range entries {
		key := classSubjectKey{e.SchoolClassID, e.SubjectID}
		lessonCounts[key] = append(lessonCounts[key], e.ID)
	}
	for key, eids := range lessonCounts {
		rKey := key.classID + "|" + key.subjectID
		req, ok := reqLookup[rKey]
		if !ok {
			continue
		}
		if len(eids) != req.LessonsPerWeek {
			c := newConflict(
				model.ConflictTooManyLessons,
				fmt.Sprintf("Class has %d lessons for %s but requires %d",
					len(eids), req.SubjectName, req.LessonsPerWeek),
				eids,
			)
			c.SchoolClassID = key.classID
			conflicts = append(conflicts, c)
		}
	}

	// ── 5. Max early starts ─────────────────────────────────────────────
	classEarlyStarts := make(map[string][]string) // classID → entry IDs with period 1
	for _, e := range entries {
		ts := slotByID[e.TimeSlotID]
		if ts != nil && ts.Period == 1 {
			classEarlyStarts[e.SchoolClassID] = append(classEarlyStarts[e.SchoolClassID], e.ID)
		}
	}
	for classID, eids := range classEarlyStarts {
		con, ok := conByClass[classID]
		if !ok {
			continue
		}
		if len(eids) > con.MaxEarlyStarts {
			c := newConflict(
				model.ConflictMaxEarlyStarts,
				fmt.Sprintf("Class %s has %d early starts but maximum is %d",
					con.SchoolClassName, len(eids), con.MaxEarlyStarts),
				eids,
			)
			c.SchoolClassID = classID
			conflicts = append(conflicts, c)
		}
	}

	// ── 6. Free afternoon violated ──────────────────────────────────────
	// For each class, check if there are afternoon entries on free afternoon days
	type classDayKey struct {
		classID string
		day     model.DayOfWeek
	}
	classAfternoonEntries := make(map[classDayKey][]string)
	for _, e := range entries {
		ts := slotByID[e.TimeSlotID]
		if ts != nil && !ts.IsMorning {
			key := classDayKey{e.SchoolClassID, ts.DayOfWeek}
			classAfternoonEntries[key] = append(classAfternoonEntries[key], e.ID)
		}
	}
	for classID, con := range conByClass {
		freeDays := parseFreeAfternoonDays(con.FreeAfternoonDays)
		for day := range freeDays {
			key := classDayKey{classID, day}
			if eids, ok := classAfternoonEntries[key]; ok && len(eids) > 0 {
				dayNames := map[model.DayOfWeek]string{
					model.Monday: "Monday", model.Tuesday: "Tuesday",
					model.Wednesday: "Wednesday", model.Thursday: "Thursday",
					model.Friday: "Friday",
				}
				c := newConflict(
					model.ConflictFreeAfternoon,
					fmt.Sprintf("Class %s has afternoon lessons on %s which should be free",
						con.SchoolClassName, dayNames[day]),
					eids,
				)
				c.SchoolClassID = classID
				conflicts = append(conflicts, c)
			}
		}
	}

	// ── 7. Max double lessons exceeded ──────────────────────────────────
	classSubjectDoubles := make(map[classSubjectKey][]string)
	for _, e := range entries {
		if e.IsDoubleLesson {
			key := classSubjectKey{e.SchoolClassID, e.SubjectID}
			classSubjectDoubles[key] = append(classSubjectDoubles[key], e.ID)
		}
	}
	for key, eids := range classSubjectDoubles {
		rKey := key.classID + "|" + key.subjectID
		req, ok := reqLookup[rKey]
		if !ok {
			continue
		}
		// Each double-lesson is represented by 2 entries; count pairs
		doublePairs := len(eids) / 2
		if doublePairs > req.MaxDoubleLessons {
			c := newConflict(
				model.ConflictMaxDoubleLessons,
				fmt.Sprintf("Subject %s has %d double-lesson pairs but maximum is %d",
					req.SubjectName, doublePairs, req.MaxDoubleLessons),
				eids,
			)
			c.SchoolClassID = key.classID
			conflicts = append(conflicts, c)
		}
	}

	return conflicts
}
