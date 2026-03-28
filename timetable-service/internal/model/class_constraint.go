package model

// ClassConstraint captures per-class scheduling constraints.
type ClassConstraint struct {
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
