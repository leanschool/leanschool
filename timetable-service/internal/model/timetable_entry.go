package model

// TimetableEntry represents a single lesson placed into the timetable grid.
type TimetableEntry struct {
	ID              string `json:"id"`
	PlanID          string `json:"planId"`
	SchoolClassID   string `json:"schoolClassId"`
	SchoolClassName string `json:"schoolClassName,omitempty"`
	SubjectID       string `json:"subjectId"`
	SubjectName     string `json:"subjectName,omitempty"`
	TeacherID       string `json:"teacherId,omitempty"`
	TeacherName     string `json:"teacherName,omitempty"`
	RoomID          string `json:"roomId,omitempty"`
	RoomName        string `json:"roomName,omitempty"`
	TimeSlotID      string `json:"timeSlotId"`
	IsDoubleLesson  bool   `json:"isDoubleLesson"`
	Version         int    `json:"version"`
}

// SwapRequest is the body for swapping two entries' time slots.
type SwapRequest struct {
	TargetEntryID string `json:"targetEntryId"`
}

// ReassignRequest is the body for reassigning a teacher to an entry.
type ReassignRequest struct {
	TeacherID string `json:"teacherId"`
}
