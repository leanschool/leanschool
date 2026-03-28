package model

// GradeRequirement specifies how many lessons per week a class needs for a subject.
type GradeRequirement struct {
	ID                string `json:"id"`
	PlanID            string `json:"planId"`
	SchoolClassID     string `json:"schoolClassId"`
	SubjectID         string `json:"subjectId"`
	SubjectName       string `json:"subjectName"`
	LessonsPerWeek    int    `json:"lessonsPerWeek"`
	MaxDoubleLessons  int    `json:"maxDoubleLessons"`
	PreferMorning     bool   `json:"preferMorning"`
	LessonDurationMin int    `json:"lessonDurationMin"`
	Version           int    `json:"version"`
}
