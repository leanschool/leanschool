//go:build integration

package integration_test

import "github.com/Joel-Haeberli/timetable-service/internal/client"

// ── Fixture IDs ───────────────────────────────────────────────────────────────
//
// IMPORTANT: Subject IDs intentionally equal subject names to work around a
// production code bug: workflow_handler.go stores subject Names in
// TeacherSnapshot.Subjects, but validator.go looks them up by SubjectID.
// Using ID==Name makes both paths work correctly.

const (
	teacherAlphaID = "teacher-alpha" // qualifies for Math and English
	teacherBetaID  = "teacher-beta"  // qualifies for History
	teacherGammaID = "teacher-gamma" // qualifies for Gym

	subjectMathID    = "Math"    // ID == Name
	subjectEnglishID = "English" // ID == Name
	subjectHistoryID = "History" // ID == Name
	subjectGymID     = "Gym"     // ID == Name

	class1aID = "class-1a"
	class2bID = "class-2b"

	room101ID = "room-101"
	room102ID = "room-102"
	gymRoomID = "room-gym"
)

// standardSnapshot returns a canonical lsMockData with 3 teachers, 4 subjects,
// 2 classes, 3 rooms.
//
// Qualification matrix:
//   - TeacherAlpha → Math, English
//   - TeacherBeta  → History
//   - TeacherGamma → Gym
func standardSnapshot() lsMockData {
	alpha := client.TeacherData{ID: teacherAlphaID, Name: "Alpha", Prename: "T"}
	beta := client.TeacherData{ID: teacherBetaID, Name: "Beta", Prename: "T"}
	gamma := client.TeacherData{ID: teacherGammaID, Name: "Gamma", Prename: "T"}

	return lsMockData{
		Teachers: []client.TeacherData{alpha, beta, gamma},
		Subjects: []client.SubjectData{
			{ID: subjectMathID, Name: subjectMathID, Teachers: []client.TeacherData{alpha}},
			{ID: subjectEnglishID, Name: subjectEnglishID, Teachers: []client.TeacherData{alpha}},
			{ID: subjectHistoryID, Name: subjectHistoryID, Teachers: []client.TeacherData{beta}},
			{ID: subjectGymID, Name: subjectGymID, Teachers: []client.TeacherData{gamma}},
		},
		Classes: []client.SchoolClassData{
			{ID: class1aID, Name: "1A", Shortcut: "1A"},
			{ID: class2bID, Name: "2B", Shortcut: "2B"},
		},
		Rooms: []client.RoomData{
			{ID: room101ID, Name: "Room 101"},
			{ID: room102ID, Name: "Room 102"},
			{ID: gymRoomID, Name: "Gym", RoomType: "gym"},
		},
	}
}

// alphaOnlyMathSnapshot returns a snapshot where only TeacherAlpha exists and
// only qualifies for Math. Useful for triggering teacher_double_booked.
func alphaOnlyMathSnapshot() lsMockData {
	alpha := client.TeacherData{ID: teacherAlphaID, Name: "Alpha", Prename: "T"}
	return lsMockData{
		Teachers: []client.TeacherData{alpha},
		Subjects: []client.SubjectData{
			{ID: subjectMathID, Name: subjectMathID, Teachers: []client.TeacherData{alpha}},
		},
		Classes: []client.SchoolClassData{
			{ID: class1aID, Name: "1A", Shortcut: "1A"},
			{ID: class2bID, Name: "2B", Shortcut: "2B"},
		},
		Rooms: []client.RoomData{
			{ID: room101ID, Name: "Room 101"},
		},
	}
}

// twoTeacherMathSnapshot returns alpha+beta both teaching Math (no double-booking
// possible with 1 slot), useful when you need qualifying teachers without conflicts.
func twoTeacherMathSnapshot() lsMockData {
	alpha := client.TeacherData{ID: teacherAlphaID, Name: "Alpha", Prename: "T"}
	beta := client.TeacherData{ID: teacherBetaID, Name: "Beta", Prename: "T"}
	return lsMockData{
		Teachers: []client.TeacherData{alpha, beta},
		Subjects: []client.SubjectData{
			{ID: subjectMathID, Name: subjectMathID, Teachers: []client.TeacherData{alpha, beta}},
		},
		Classes: []client.SchoolClassData{
			{ID: class1aID, Name: "1A", Shortcut: "1A"},
			{ID: class2bID, Name: "2B", Shortcut: "2B"},
		},
		Rooms: []client.RoomData{
			{ID: room101ID, Name: "Room 101"},
		},
	}
}
