package model

// TeacherSnapshot is a plan-local copy of a teacher from the leanschool API.
type TeacherSnapshot struct {
	ID       string   `json:"id"`
	PlanID   string   `json:"planId"`
	Name     string   `json:"name"`
	Prename  string   `json:"prename"`
	Sub      string   `json:"sub,omitempty"`
	Subjects []string `json:"subjects"`
}

// SubjectSnapshot is a plan-local copy of a subject.
type SubjectSnapshot struct {
	ID     string `json:"id"`
	PlanID string `json:"planId"`
	Name   string `json:"name"`
}

// SchoolClassSnapshot is a plan-local copy of a school class.
type SchoolClassSnapshot struct {
	ID       string `json:"id"`
	PlanID   string `json:"planId"`
	Name     string `json:"name"`
	Shortcut string `json:"shortcut,omitempty"`
}

// RoomSnapshot is a plan-local copy of a room.
type RoomSnapshot struct {
	ID       string `json:"id"`
	PlanID   string `json:"planId"`
	Name     string `json:"name"`
	RoomType string `json:"roomType,omitempty"`
}
