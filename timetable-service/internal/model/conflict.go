package model

// ConflictType categorises scheduling conflicts.
type ConflictType string

const (
	ConflictTeacherDoubleBooked ConflictType = "teacher_double_booked"
	ConflictRoomDoubleBooked    ConflictType = "room_double_booked"
	ConflictTooManyLessons      ConflictType = "too_many_lessons"
	ConflictWrongTimeBlock      ConflictType = "wrong_time_block"
	ConflictTeacherNotQualified ConflictType = "teacher_not_qualified"
	ConflictMaxEarlyStarts      ConflictType = "max_early_starts_exceeded"
	ConflictFreeAfternoon       ConflictType = "free_afternoon_violated"
	ConflictMaxDoubleLessons    ConflictType = "max_double_lessons_exceeded"
)

// ConflictSeverity indicates how serious a conflict is.
type ConflictSeverity string

const (
	SeverityError   ConflictSeverity = "error"
	SeverityWarning ConflictSeverity = "warning"
)

// Conflict describes a single scheduling conflict detected during validation.
type Conflict struct {
	ID            string           `json:"id"`
	PlanID        string           `json:"planId"`
	Type          ConflictType     `json:"type"`
	Severity      ConflictSeverity `json:"severity"`
	Description   string           `json:"description"`
	EntryIDs      []string         `json:"entryIds"`
	TeacherID     string           `json:"teacherId,omitempty"`
	SchoolClassID string           `json:"schoolClassId,omitempty"`
	TimeSlotID    string           `json:"timeSlotId,omitempty"`
	Resolved      bool             `json:"resolved"`
	ResolvedBy    string           `json:"resolvedBy,omitempty"`
	Version       int              `json:"version"`
}
