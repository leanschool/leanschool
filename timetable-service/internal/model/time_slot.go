package model

// DayOfWeek represents a weekday (1=Monday through 5=Friday).
type DayOfWeek int

const (
	Monday    DayOfWeek = 1
	Tuesday   DayOfWeek = 2
	Wednesday DayOfWeek = 3
	Thursday  DayOfWeek = 4
	Friday    DayOfWeek = 5
)

// TimeSlotDefinition defines a single period in the weekly grid.
type TimeSlotDefinition struct {
	ID        string    `json:"id"`
	PlanID    string    `json:"planId"`
	DayOfWeek DayOfWeek `json:"dayOfWeek"`
	Period    int       `json:"period"`
	StartTime string    `json:"startTime"`
	EndTime   string    `json:"endTime"`
	IsMorning bool      `json:"isMorning"`
	Version   int       `json:"version"`
}

// GenerateDefaultRequest is the body for the generate-default endpoint.
type GenerateDefaultRequest struct {
	MorningPeriods    int    `json:"morningPeriods"`
	AfternoonPeriods  int    `json:"afternoonPeriods"`
	StartTime         string `json:"startTime"`
	LessonDurationMin int    `json:"lessonDurationMin"`
	BreakDurationMin  int    `json:"breakDurationMin"`
	LunchBreakMin     int    `json:"lunchBreakMin"`
}
