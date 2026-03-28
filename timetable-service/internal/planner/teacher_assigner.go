package planner

import (
	"sort"

	"github.com/Joel-Haeberli/timetable-service/internal/model"
)

// AssignTeachers assigns teachers to timetable entries based on subject
// qualifications and availability. It uses a greedy load-balancing strategy:
// for each entry, the qualified teacher with the fewest existing assignments
// at the same time slot is preferred.
func AssignTeachers(
	entries []*model.TimetableEntry,
	teachers []*model.TeacherSnapshot,
	timeSlots []*model.TimeSlotDefinition,
) {
	// Build subject → qualified teachers index
	subjectTeachers := make(map[string][]*model.TeacherSnapshot)
	for _, t := range teachers {
		for _, sid := range t.Subjects {
			subjectTeachers[sid] = append(subjectTeachers[sid], t)
		}
	}

	// Sort entries by constraint difficulty: subjects with fewer qualified
	// teachers should be assigned first (most constrained first).
	sort.SliceStable(entries, func(i, j int) bool {
		ci := len(subjectTeachers[entries[i].SubjectID])
		cj := len(subjectTeachers[entries[j].SubjectID])
		return ci < cj
	})

	// Track teacher assignments: teacherID → set of timeSlotIDs already assigned
	teacherSlots := make(map[string]map[string]bool)

	// Build time slot lookup for finding consecutive periods (double-lessons)
	slotByID := make(map[string]*model.TimeSlotDefinition, len(timeSlots))
	for _, ts := range timeSlots {
		slotByID[ts.ID] = ts
	}

	for _, entry := range entries {
		qualified := subjectTeachers[entry.SubjectID]
		if len(qualified) == 0 {
			continue // no qualified teacher; validator will flag this
		}

		var bestTeacher *model.TeacherSnapshot
		bestConflicts := -1

		for _, t := range qualified {
			slots, ok := teacherSlots[t.ID]
			if !ok {
				slots = make(map[string]bool)
			}

			conflicts := 0
			if slots[entry.TimeSlotID] {
				conflicts++
			}

			// For double-lessons, also check the next period
			if entry.IsDoubleLesson {
				if ts, ok := slotByID[entry.TimeSlotID]; ok {
					nextSlot := findNextPeriodSlot(ts, buildTimeGrid(timeSlots))
					if nextSlot != nil && slots[nextSlot.ID] {
						conflicts++
					}
				}
			}

			if bestTeacher == nil || conflicts < bestConflicts {
				bestTeacher = t
				bestConflicts = conflicts
			}
		}

		if bestTeacher != nil {
			entry.TeacherID = bestTeacher.ID
			entry.TeacherName = bestTeacher.Prename + " " + bestTeacher.Name

			if teacherSlots[bestTeacher.ID] == nil {
				teacherSlots[bestTeacher.ID] = make(map[string]bool)
			}
			teacherSlots[bestTeacher.ID][entry.TimeSlotID] = true

			// For double-lessons, also mark the next period slot
			if entry.IsDoubleLesson {
				if ts, ok := slotByID[entry.TimeSlotID]; ok {
					nextSlot := findNextPeriodSlot(ts, buildTimeGrid(timeSlots))
					if nextSlot != nil {
						teacherSlots[bestTeacher.ID][nextSlot.ID] = true
					}
				}
			}
		}
	}
}
