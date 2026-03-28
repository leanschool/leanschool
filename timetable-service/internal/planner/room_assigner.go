package planner

import "github.com/Joel-Haeberli/timetable-service/internal/model"

// AssignRooms is a placeholder for room assignment logic. In V1, room
// assignments are left empty and can be done manually by planners.
// A future iteration may assign each class's default classroom or use a
// round-robin strategy across available rooms.
func AssignRooms(
	entries []*model.TimetableEntry,
	classes []*model.SchoolClassSnapshot,
	rooms []*model.RoomSnapshot,
) {
	// V1: no automatic room assignment.
	// Rooms can be assigned manually through the entry update endpoint.
}
