package storage

import (
	"context"

	model "github.com/Joel-Haeberli/leanschool-model"
)

// Storage defines CRUD operations for all domain entities.
type Storage interface {
	// receipts
	Create(ctx context.Context, r *model.Receipt) error
	Get(ctx context.Context, id string) (*model.Receipt, error)
	List(ctx context.Context, ownerID string) ([]*model.Receipt, error)
	Update(ctx context.Context, r *model.Receipt) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, ids []string, status model.ReceiptStatus) error

	// accounts
	CreateAccount(ctx context.Context, a *model.Account) error
	GetAccount(ctx context.Context, id string) (*model.Account, error)
	ListAccounts(ctx context.Context) ([]*model.Account, error)
	UpdateAccount(ctx context.Context, a *model.Account) error
	DeleteAccount(ctx context.Context, id string) error

	// school classes
	CreateSchoolClass(ctx context.Context, sc *model.SchoolClass) error
	GetSchoolClass(ctx context.Context, id string) (*model.SchoolClass, error)
	ListSchoolClasses(ctx context.Context) ([]*model.SchoolClass, error)
	UpdateSchoolClass(ctx context.Context, sc *model.SchoolClass) error
	DeleteSchoolClass(ctx context.Context, id string) error

	// registration requests
	CreateRegistrationRequest(ctx context.Context, req *model.RegistrationRequest) error
	GetRegistrationRequestByUserSub(ctx context.Context, userSub string) (*model.RegistrationRequest, error)
	GetRegistrationRequestByID(ctx context.Context, id string) (*model.RegistrationRequest, error)
	ListRegistrationRequests(ctx context.Context) ([]*model.RegistrationRequest, error)
	UpdateRegistrationRequestStatus(ctx context.Context, id string, status model.RegistrationStatus) error

	// user profiles
	GetUserProfile(ctx context.Context, userSub string) (*model.UserProfile, error)
	UpsertUserProfile(ctx context.Context, profile *model.UserProfile) error

	// user registry
	CreateUserRegistry(ctx context.Context, user *model.UserRegistry) error
	GetUserRegistry(ctx context.Context, id string) (*model.UserRegistry, error)
	GetUserRegistryBySub(ctx context.Context, userSub string) (*model.UserRegistry, error)
	GetUserRegistryByPersonID(ctx context.Context, personID string) (*model.UserRegistry, error)
	UpdateUserRegistry(ctx context.Context, user *model.UserRegistry) error
	ListUserRegistries(ctx context.Context) ([]*model.UserRegistry, error)

	// registration workflow
	CreateRegistrationWorkflow(ctx context.Context, workflow *model.RegistrationWorkflow) error
	GetRegistrationWorkflow(ctx context.Context, id string) (*model.RegistrationWorkflow, error)
	ListRegistrationWorkflows(ctx context.Context, statusFilter string) ([]*model.RegistrationWorkflow, error)
	UpdateRegistrationWorkflow(ctx context.Context, workflow *model.RegistrationWorkflow) error

	// role mappings
	CreateRoleMapping(ctx context.Context, mapping *model.RoleMapping) error
	GetRoleMapping(ctx context.Context, keycloakRole string) (*model.RoleMapping, error)
	ListRoleMappings(ctx context.Context) ([]*model.RoleMapping, error)
	UpdateRoleMapping(ctx context.Context, mapping *model.RoleMapping) error
	DeleteRoleMapping(ctx context.Context, keycloakRole string) error

	// user linking
	EnsureUserLinked(ctx context.Context, userSub string, personData *model.Person) error
	SyncUserRoles(ctx context.Context, userSub string) error

	// locations
	CreateLocation(ctx context.Context, l *model.Location) error
	GetLocation(ctx context.Context, id string) (*model.Location, error)
	ListLocations(ctx context.Context) ([]*model.Location, error)
	UpdateLocation(ctx context.Context, l *model.Location) error
	DeleteLocation(ctx context.Context, id string) error

	// buildings
	CreateBuilding(ctx context.Context, b *model.Building) error
	GetBuilding(ctx context.Context, id string) (*model.Building, error)
	ListBuildings(ctx context.Context) ([]*model.Building, error)
	UpdateBuilding(ctx context.Context, b *model.Building) error
	DeleteBuilding(ctx context.Context, id string) error

	// rooms
	CreateRoom(ctx context.Context, r *model.Room) error
	GetRoom(ctx context.Context, id string) (*model.Room, error)
	ListRooms(ctx context.Context) ([]*model.Room, error)
	UpdateRoom(ctx context.Context, r *model.Room) error
	DeleteRoom(ctx context.Context, id string) error

	// postal codes
	CreatePostalCode(ctx context.Context, pc *model.PostalCode) error
	GetPostalCode(ctx context.Context, id string) (*model.PostalCode, error)
	ListPostalCodes(ctx context.Context) ([]*model.PostalCode, error)
	UpdatePostalCode(ctx context.Context, pc *model.PostalCode) error
	DeletePostalCode(ctx context.Context, id string) error

	// cities
	CreateCity(ctx context.Context, c *model.City) error
	GetCity(ctx context.Context, id string) (*model.City, error)
	ListCities(ctx context.Context) ([]*model.City, error)
	UpdateCity(ctx context.Context, c *model.City) error
	DeleteCity(ctx context.Context, id string) error

	// addresses
	CreateAddress(ctx context.Context, a *model.Address) error
	GetAddress(ctx context.Context, id string) (*model.Address, error)
	ListAddresses(ctx context.Context) ([]*model.Address, error)
	UpdateAddress(ctx context.Context, a *model.Address) error
	DeleteAddress(ctx context.Context, id string) error

	// persons
	CreatePerson(ctx context.Context, p *model.Person) error
	GetPerson(ctx context.Context, id string) (*model.Person, error)
	ListPersons(ctx context.Context) ([]*model.Person, error)
	UpdatePerson(ctx context.Context, p *model.Person) error
	DeletePerson(ctx context.Context, id string) error

	// guardians
	CreateGuardian(ctx context.Context, g *model.Guardian) error
	GetGuardian(ctx context.Context, id string) (*model.Guardian, error)
	ListGuardians(ctx context.Context) ([]*model.Guardian, error)
	UpdateGuardian(ctx context.Context, g *model.Guardian) error
	DeleteGuardian(ctx context.Context, id string) error

	// teachers
	CreateTeacher(ctx context.Context, t *model.Teacher) error
	GetTeacher(ctx context.Context, id string) (*model.Teacher, error)
	ListTeachers(ctx context.Context) ([]*model.Teacher, error)
	UpdateTeacher(ctx context.Context, t *model.Teacher) error
	DeleteTeacher(ctx context.Context, id string) error

	// students
	CreateStudent(ctx context.Context, s *model.Student) error
	GetStudent(ctx context.Context, id string) (*model.Student, error)
	ListStudents(ctx context.Context) ([]*model.Student, error)
	UpdateStudent(ctx context.Context, s *model.Student) error
	DeleteStudent(ctx context.Context, id string) error

	// school years
	CreateSchoolYear(ctx context.Context, sy *model.SchoolYear) error
	GetSchoolYear(ctx context.Context, id string) (*model.SchoolYear, error)
	ListSchoolYears(ctx context.Context) ([]*model.SchoolYear, error)
	UpdateSchoolYear(ctx context.Context, sy *model.SchoolYear) error
	DeleteSchoolYear(ctx context.Context, id string) error

	// curricula
	CreateCurriculum(ctx context.Context, c *model.Curriculum) error
	GetCurriculum(ctx context.Context, id string) (*model.Curriculum, error)
	ListCurricula(ctx context.Context) ([]*model.Curriculum, error)
	UpdateCurriculum(ctx context.Context, c *model.Curriculum) error
	DeleteCurriculum(ctx context.Context, id string) error

	// subjects
	CreateSubject(ctx context.Context, s *model.Subject) error
	GetSubject(ctx context.Context, id string) (*model.Subject, error)
	ListSubjects(ctx context.Context) ([]*model.Subject, error)
	UpdateSubject(ctx context.Context, s *model.Subject) error
	DeleteSubject(ctx context.Context, id string) error

	// lessons
	CreateLesson(ctx context.Context, l *model.Lesson) error
	GetLesson(ctx context.Context, id string) (*model.Lesson, error)
	ListLessons(ctx context.Context) ([]*model.Lesson, error)
	UpdateLesson(ctx context.Context, l *model.Lesson) error
	DeleteLesson(ctx context.Context, id string) error

	// exams
	CreateExam(ctx context.Context, e *model.Exam) error
	GetExam(ctx context.Context, id string) (*model.Exam, error)
	ListExams(ctx context.Context) ([]*model.Exam, error)
	UpdateExam(ctx context.Context, e *model.Exam) error
	DeleteExam(ctx context.Context, id string) error
	// IsTeacherOfExam returns true if the user identified by sub is a teacher of the school class associated with examID.
	IsTeacherOfExam(ctx context.Context, examID, sub string) (bool, error)

	// grades
	CreateGrade(ctx context.Context, g *model.Grade) error
	GetGrade(ctx context.Context, id string) (*model.Grade, error)
	ListGrades(ctx context.Context) ([]*model.Grade, error)
	UpdateGrade(ctx context.Context, g *model.Grade) error
	DeleteGrade(ctx context.Context, id string) error
	// IsTeacherOfGrade returns true if the user identified by sub is a teacher of the school class associated with the grade's exam.
	IsTeacherOfGrade(ctx context.Context, gradeID, sub string) (bool, error)
}
