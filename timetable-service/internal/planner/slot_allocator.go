package planner

import (
	"fmt"
	"sort"

	"github.com/Joel-Haeberli/timetable-service/internal/model"
)

// lessonNeed represents one slot that must be placed for a subject in a class.
type lessonNeed struct {
	schoolClassID   string
	schoolClassName string
	subjectID       string
	subjectName     string
	preferMorning   bool
	lessonsPerWeek  int // used for sort priority
	isDoublePart    bool
	doublePartner   *lessonNeed // paired double-lesson partner
}

// AllocateSlots distributes subject lessons across the weekly time grid for
// every class that has HasTimetable==true. It returns a slice of
// TimetableEntry with PlanID, SchoolClassID, SubjectID, TimeSlotID, and
// IsDoubleLesson populated. TeacherID and RoomID are left empty.
func AllocateSlots(
	requirements []*model.GradeRequirement,
	constraints []*model.ClassConstraint,
	timeSlots []*model.TimeSlotDefinition,
) []*model.TimetableEntry {

	// ── index data ──────────────────────────────────────────────────────

	// requirements grouped by class
	reqByClass := make(map[string][]*model.GradeRequirement)
	for _, r := range requirements {
		reqByClass[r.SchoolClassID] = append(reqByClass[r.SchoolClassID], r)
	}

	// constraints indexed by class
	conByClass := make(map[string]*model.ClassConstraint)
	for _, c := range constraints {
		conByClass[c.SchoolClassID] = c
	}

	// time grid: day → sorted periods
	grid := buildTimeGrid(timeSlots)

	// slot lookup for quick access
	slotByID := make(map[string]*model.TimeSlotDefinition, len(timeSlots))
	for _, ts := range timeSlots {
		slotByID[ts.ID] = ts
	}

	// ── per-class allocation ────────────────────────────────────────────

	var entries []*model.TimetableEntry
	entryCounter := 0

	for _, con := range constraints {
		if !con.HasTimetable {
			continue
		}
		classID := con.SchoolClassID
		className := con.SchoolClassName
		reqs := reqByClass[classID]
		if len(reqs) == 0 {
			continue
		}

		// Build lesson needs
		needs := buildLessonNeeds(reqs, className)

		// Sort: preferMorning first, then harder-to-place (more lessons) first
		sort.SliceStable(needs, func(i, j int) bool {
			if needs[i].preferMorning != needs[j].preferMorning {
				return needs[i].preferMorning
			}
			return needs[i].lessonsPerWeek > needs[j].lessonsPerWeek
		})

		// Parse free afternoon days
		freeDays := parseFreeAfternoonDays(con.FreeAfternoonDays)

		// Tracking
		usedSlots := make(map[string]bool)
		earlyStartCount := 0

		for _, need := range needs {
			// skip double partner — it's placed with its pair
			if need.isDoublePart && need.doublePartner != nil {
				continue
			}

			placed := false

			if need.doublePartner != nil {
				// Need two consecutive periods on the same day
				placed = placeDoubleLesson(
					need, grid, slotByID, con, freeDays,
					usedSlots, &earlyStartCount, &entries, &entryCounter,
				)
			}

			if !placed && need.doublePartner == nil {
				placed = placeSingleLesson(
					need, grid, slotByID, con, freeDays,
					usedSlots, &earlyStartCount, &entries, &entryCounter,
				)
			}

			// Fallback: place in any available slot (conflict caught by validator)
			if !placed {
				if need.doublePartner != nil {
					placeFallbackDouble(
						need, grid, slotByID,
						usedSlots, &entries, &entryCounter,
					)
				} else {
					placeFallbackSingle(
						need, grid, slotByID,
						usedSlots, &entries, &entryCounter,
					)
				}
			}
		}
	}

	return entries
}

// buildTimeGrid constructs a map from DayOfWeek to sorted time slot definitions.
func buildTimeGrid(timeSlots []*model.TimeSlotDefinition) map[model.DayOfWeek][]*model.TimeSlotDefinition {
	grid := make(map[model.DayOfWeek][]*model.TimeSlotDefinition)
	for _, ts := range timeSlots {
		grid[ts.DayOfWeek] = append(grid[ts.DayOfWeek], ts)
	}
	for day := range grid {
		sort.Slice(grid[day], func(i, j int) bool {
			return grid[day][i].Period < grid[day][j].Period
		})
	}
	return grid
}

// buildLessonNeeds expands requirements into individual lesson needs,
// pairing double-lessons where allowed.
func buildLessonNeeds(reqs []*model.GradeRequirement, className string) []*lessonNeed {
	var needs []*lessonNeed

	for _, req := range reqs {
		remaining := req.LessonsPerWeek
		doubles := req.MaxDoubleLessons

		// Create double-lesson pairs
		for doubles > 0 && remaining >= 2 {
			a := &lessonNeed{
				schoolClassID:   req.SchoolClassID,
				schoolClassName: className,
				subjectID:       req.SubjectID,
				subjectName:     req.SubjectName,
				preferMorning:   req.PreferMorning,
				lessonsPerWeek:  req.LessonsPerWeek,
				isDoublePart:    false,
			}
			b := &lessonNeed{
				schoolClassID:   req.SchoolClassID,
				schoolClassName: className,
				subjectID:       req.SubjectID,
				subjectName:     req.SubjectName,
				preferMorning:   req.PreferMorning,
				lessonsPerWeek:  req.LessonsPerWeek,
				isDoublePart:    true,
			}
			a.doublePartner = b
			b.doublePartner = a
			needs = append(needs, a, b)
			remaining -= 2
			doubles--
		}

		// Create single-lesson needs for the rest
		for i := 0; i < remaining; i++ {
			needs = append(needs, &lessonNeed{
				schoolClassID:   req.SchoolClassID,
				schoolClassName: className,
				subjectID:       req.SubjectID,
				subjectName:     req.SubjectName,
				preferMorning:   req.PreferMorning,
				lessonsPerWeek:  req.LessonsPerWeek,
			})
		}
	}

	return needs
}

// parseFreeAfternoonDays parses a comma-separated string of day numbers
// (e.g. "3,5") into a set of DayOfWeek values.
func parseFreeAfternoonDays(s string) map[model.DayOfWeek]bool {
	result := make(map[model.DayOfWeek]bool)
	if s == "" {
		return result
	}
	for _, ch := range s {
		if ch >= '1' && ch <= '5' {
			result[model.DayOfWeek(ch-'0')] = true
		}
	}
	return result
}

// slotKey returns a unique key for a class-slot combination.
func slotKey(classID, slotID string) string {
	return classID + "|" + slotID
}

// isSlotValid checks whether a time slot can be used for a class given constraints.
func isSlotValid(
	ts *model.TimeSlotDefinition,
	con *model.ClassConstraint,
	freeDays map[model.DayOfWeek]bool,
	usedSlots map[string]bool,
	classID string,
	earlyStarts int,
) bool {
	if usedSlots[slotKey(classID, ts.ID)] {
		return false
	}
	// Free afternoon check: no afternoon slots on free days
	if !ts.IsMorning && freeDays[ts.DayOfWeek] {
		return false
	}
	// Max early starts check (period 1)
	if ts.Period == 1 && earlyStarts >= con.MaxEarlyStarts {
		return false
	}
	return true
}

// findNextPeriodSlot finds the time slot with period = ts.Period+1 on the same day.
func findNextPeriodSlot(
	ts *model.TimeSlotDefinition,
	grid map[model.DayOfWeek][]*model.TimeSlotDefinition,
) *model.TimeSlotDefinition {
	for _, s := range grid[ts.DayOfWeek] {
		if s.Period == ts.Period+1 {
			return s
		}
	}
	return nil
}

func makeEntry(need *lessonNeed, slotID string, isDouble bool, counter *int) *model.TimetableEntry {
	*counter++
	return &model.TimetableEntry{
		ID:              fmt.Sprintf("gen-%d", *counter),
		SchoolClassID:   need.schoolClassID,
		SchoolClassName: need.schoolClassName,
		SubjectID:       need.subjectID,
		SubjectName:     need.subjectName,
		TimeSlotID:      slotID,
		IsDoubleLesson:  isDouble,
	}
}

func placeDoubleLesson(
	need *lessonNeed,
	grid map[model.DayOfWeek][]*model.TimeSlotDefinition,
	slotByID map[string]*model.TimeSlotDefinition,
	con *model.ClassConstraint,
	freeDays map[model.DayOfWeek]bool,
	usedSlots map[string]bool,
	earlyStarts *int,
	entries *[]*model.TimetableEntry,
	counter *int,
) bool {
	// Order days to try: all weekdays
	days := []model.DayOfWeek{model.Monday, model.Tuesday, model.Wednesday, model.Thursday, model.Friday}

	// If preferMorning, try morning slots first within each day
	for _, day := range days {
		slots := grid[day]
		for i := 0; i < len(slots)-1; i++ {
			ts1 := slots[i]
			ts2 := slots[i+1]
			if ts2.Period != ts1.Period+1 {
				continue
			}

			// For preferMorning, skip if not morning on first pass
			if need.preferMorning && !ts1.IsMorning {
				continue
			}

			if !isSlotValid(ts1, con, freeDays, usedSlots, need.schoolClassID, *earlyStarts) {
				continue
			}
			if !isSlotValid(ts2, con, freeDays, usedSlots, need.schoolClassID, *earlyStarts) {
				continue
			}

			// Place
			e1 := makeEntry(need, ts1.ID, true, counter)
			e2 := makeEntry(need, ts2.ID, true, counter)
			*entries = append(*entries, e1, e2)
			usedSlots[slotKey(need.schoolClassID, ts1.ID)] = true
			usedSlots[slotKey(need.schoolClassID, ts2.ID)] = true
			if ts1.Period == 1 {
				*earlyStarts++
			}
			return true
		}
	}

	// Second pass: if preferMorning was set, try afternoon too
	if need.preferMorning {
		for _, day := range days {
			slots := grid[day]
			for i := 0; i < len(slots)-1; i++ {
				ts1 := slots[i]
				ts2 := slots[i+1]
				if ts2.Period != ts1.Period+1 {
					continue
				}
				if ts1.IsMorning {
					continue // already tried these
				}
				if !isSlotValid(ts1, con, freeDays, usedSlots, need.schoolClassID, *earlyStarts) {
					continue
				}
				if !isSlotValid(ts2, con, freeDays, usedSlots, need.schoolClassID, *earlyStarts) {
					continue
				}

				e1 := makeEntry(need, ts1.ID, true, counter)
				e2 := makeEntry(need, ts2.ID, true, counter)
				*entries = append(*entries, e1, e2)
				usedSlots[slotKey(need.schoolClassID, ts1.ID)] = true
				usedSlots[slotKey(need.schoolClassID, ts2.ID)] = true
				return true
			}
		}
	}

	return false
}

func placeSingleLesson(
	need *lessonNeed,
	grid map[model.DayOfWeek][]*model.TimeSlotDefinition,
	slotByID map[string]*model.TimeSlotDefinition,
	con *model.ClassConstraint,
	freeDays map[model.DayOfWeek]bool,
	usedSlots map[string]bool,
	earlyStarts *int,
	entries *[]*model.TimetableEntry,
	counter *int,
) bool {
	days := []model.DayOfWeek{model.Monday, model.Tuesday, model.Wednesday, model.Thursday, model.Friday}

	// First pass: try preferred slots (morning if preferMorning)
	for _, day := range days {
		for _, ts := range grid[day] {
			if need.preferMorning && !ts.IsMorning {
				continue
			}
			if !isSlotValid(ts, con, freeDays, usedSlots, need.schoolClassID, *earlyStarts) {
				continue
			}
			e := makeEntry(need, ts.ID, false, counter)
			*entries = append(*entries, e)
			usedSlots[slotKey(need.schoolClassID, ts.ID)] = true
			if ts.Period == 1 {
				*earlyStarts++
			}
			return true
		}
	}

	// Second pass: try all remaining slots
	if need.preferMorning {
		for _, day := range days {
			for _, ts := range grid[day] {
				if ts.IsMorning {
					continue // already tried
				}
				if !isSlotValid(ts, con, freeDays, usedSlots, need.schoolClassID, *earlyStarts) {
					continue
				}
				e := makeEntry(need, ts.ID, false, counter)
				*entries = append(*entries, e)
				usedSlots[slotKey(need.schoolClassID, ts.ID)] = true
				return true
			}
		}
	}

	return false
}

func placeFallbackDouble(
	need *lessonNeed,
	grid map[model.DayOfWeek][]*model.TimeSlotDefinition,
	slotByID map[string]*model.TimeSlotDefinition,
	usedSlots map[string]bool,
	entries *[]*model.TimetableEntry,
	counter *int,
) {
	days := []model.DayOfWeek{model.Monday, model.Tuesday, model.Wednesday, model.Thursday, model.Friday}
	for _, day := range days {
		slots := grid[day]
		for i := 0; i < len(slots)-1; i++ {
			ts1 := slots[i]
			ts2 := slots[i+1]
			if ts2.Period != ts1.Period+1 {
				continue
			}
			k1 := slotKey(need.schoolClassID, ts1.ID)
			k2 := slotKey(need.schoolClassID, ts2.ID)
			if usedSlots[k1] || usedSlots[k2] {
				continue
			}
			e1 := makeEntry(need, ts1.ID, true, counter)
			e2 := makeEntry(need, ts2.ID, true, counter)
			*entries = append(*entries, e1, e2)
			usedSlots[k1] = true
			usedSlots[k2] = true
			return
		}
	}
	// Last resort: place as two singles in any free slots
	placed := 0
	for _, day := range days {
		for _, ts := range grid[day] {
			k := slotKey(need.schoolClassID, ts.ID)
			if usedSlots[k] {
				continue
			}
			e := makeEntry(need, ts.ID, true, counter)
			*entries = append(*entries, e)
			usedSlots[k] = true
			placed++
			if placed >= 2 {
				return
			}
		}
	}
}

func placeFallbackSingle(
	need *lessonNeed,
	grid map[model.DayOfWeek][]*model.TimeSlotDefinition,
	slotByID map[string]*model.TimeSlotDefinition,
	usedSlots map[string]bool,
	entries *[]*model.TimetableEntry,
	counter *int,
) {
	days := []model.DayOfWeek{model.Monday, model.Tuesday, model.Wednesday, model.Thursday, model.Friday}
	for _, day := range days {
		for _, ts := range grid[day] {
			k := slotKey(need.schoolClassID, ts.ID)
			if usedSlots[k] {
				continue
			}
			e := makeEntry(need, ts.ID, false, counter)
			*entries = append(*entries, e)
			usedSlots[k] = true
			return
		}
	}
}
