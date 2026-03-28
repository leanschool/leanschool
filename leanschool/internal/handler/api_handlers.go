package handler

import (
	"context"
	"encoding/json"
	"errors"

	api "github.com/Joel-Haeberli/leanschool/internal/api"
	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

// Handlers implements api.StrictServerInterface using the storage layer.
type Handlers struct {
	store        storage.Storage
	templatePath string
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(store storage.Storage, templatePath string) *Handlers {
	return &Handlers{store: store, templatePath: templatePath}
}

// jsonConvert marshals src to JSON and unmarshals into dst.
// This works because api.* and model.* types share identical JSON tags.
func jsonConvert(src, dst any) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

// ── Addresses ────────────────────────────────────────────────────────────────

func (h *Handlers) ListAddresss(ctx context.Context, req api.ListAddresssRequestObject) (api.ListAddresssResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "address_read") {
		return api.ListAddresss403Response{}, nil
	}
	list, err := h.store.ListAddresses(ctx)
	if err != nil {
		return nil, err
	}
	var resp []api.Address
	jsonConvert(list, &resp)
	if resp == nil {
		resp = []api.Address{}
	}
	return api.ListAddresss200JSONResponse(resp), nil
}

func (h *Handlers) CreateAddress(ctx context.Context, req api.CreateAddressRequestObject) (api.CreateAddressResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "address_write") {
		return api.CreateAddress403Response{}, nil
	}
	var m model.Address
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.CreateAddress400Response{}, nil
	}
	if err := h.store.CreateAddress(ctx, &m); err != nil {
		return api.CreateAddress400Response{}, nil
	}
	var resp api.Address
	jsonConvert(&m, &resp)
	return api.CreateAddress201JSONResponse(resp), nil
}

func (h *Handlers) DeleteAddress(ctx context.Context, req api.DeleteAddressRequestObject) (api.DeleteAddressResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "address_write") {
		return api.DeleteAddress403Response{}, nil
	}
	if err := h.store.DeleteAddress(ctx, req.ID); err != nil {
		return nil, err
	}
	return api.DeleteAddress204Response{}, nil
}

func (h *Handlers) GetAddress(ctx context.Context, req api.GetAddressRequestObject) (api.GetAddressResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "address_read") {
		return api.GetAddress404Response{}, nil
	}
	m, err := h.store.GetAddress(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return api.GetAddress404Response{}, nil
	}
	var resp api.Address
	jsonConvert(m, &resp)
	return api.GetAddress200JSONResponse(resp), nil
}

func (h *Handlers) UpdateAddress(ctx context.Context, req api.UpdateAddressRequestObject) (api.UpdateAddressResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "address_write") {
		return api.UpdateAddress403Response{}, nil
	}
	var m model.Address
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.UpdateAddress400Response{}, nil
	}
	m.ID = req.ID
	if err := h.store.UpdateAddress(ctx, &m); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			return api.UpdateAddress409Response{}, nil
		}
		return api.UpdateAddress400Response{}, nil
	}
	var resp api.Address
	jsonConvert(&m, &resp)
	return api.UpdateAddress200JSONResponse(resp), nil
}

// ── Buildings ─────────────────────────────────────────────────────────────────

func (h *Handlers) ListBuildings(ctx context.Context, req api.ListBuildingsRequestObject) (api.ListBuildingsResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "building_read") {
		return api.ListBuildings403Response{}, nil
	}
	list, err := h.store.ListBuildings(ctx)
	if err != nil {
		return nil, err
	}
	var resp []api.Building
	jsonConvert(list, &resp)
	if resp == nil {
		resp = []api.Building{}
	}
	return api.ListBuildings200JSONResponse(resp), nil
}

func (h *Handlers) CreateBuilding(ctx context.Context, req api.CreateBuildingRequestObject) (api.CreateBuildingResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "building_write") {
		return api.CreateBuilding403Response{}, nil
	}
	var m model.Building
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.CreateBuilding400Response{}, nil
	}
	if err := h.store.CreateBuilding(ctx, &m); err != nil {
		return api.CreateBuilding400Response{}, nil
	}
	var resp api.Building
	jsonConvert(&m, &resp)
	return api.CreateBuilding201JSONResponse(resp), nil
}

func (h *Handlers) DeleteBuilding(ctx context.Context, req api.DeleteBuildingRequestObject) (api.DeleteBuildingResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "building_write") {
		return api.DeleteBuilding403Response{}, nil
	}
	if err := h.store.DeleteBuilding(ctx, req.ID); err != nil {
		return nil, err
	}
	return api.DeleteBuilding204Response{}, nil
}

func (h *Handlers) GetBuilding(ctx context.Context, req api.GetBuildingRequestObject) (api.GetBuildingResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "building_read") {
		return api.GetBuilding404Response{}, nil
	}
	m, err := h.store.GetBuilding(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return api.GetBuilding404Response{}, nil
	}
	var resp api.Building
	jsonConvert(m, &resp)
	return api.GetBuilding200JSONResponse(resp), nil
}

func (h *Handlers) UpdateBuilding(ctx context.Context, req api.UpdateBuildingRequestObject) (api.UpdateBuildingResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "building_write") {
		return api.UpdateBuilding403Response{}, nil
	}
	var m model.Building
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.UpdateBuilding400Response{}, nil
	}
	m.ID = req.ID
	if err := h.store.UpdateBuilding(ctx, &m); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			return api.UpdateBuilding409Response{}, nil
		}
		return api.UpdateBuilding400Response{}, nil
	}
	var resp api.Building
	jsonConvert(&m, &resp)
	return api.UpdateBuilding200JSONResponse(resp), nil
}

// ── Cities ────────────────────────────────────────────────────────────────────

func (h *Handlers) ListCitys(ctx context.Context, req api.ListCitysRequestObject) (api.ListCitysResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "city_read") {
		return api.ListCitys403Response{}, nil
	}
	list, err := h.store.ListCities(ctx)
	if err != nil {
		return nil, err
	}
	var resp []api.City
	jsonConvert(list, &resp)
	if resp == nil {
		resp = []api.City{}
	}
	return api.ListCitys200JSONResponse(resp), nil
}

func (h *Handlers) CreateCity(ctx context.Context, req api.CreateCityRequestObject) (api.CreateCityResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "city_write") {
		return api.CreateCity403Response{}, nil
	}
	var m model.City
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.CreateCity400Response{}, nil
	}
	if err := h.store.CreateCity(ctx, &m); err != nil {
		return api.CreateCity400Response{}, nil
	}
	var resp api.City
	jsonConvert(&m, &resp)
	return api.CreateCity201JSONResponse(resp), nil
}

func (h *Handlers) DeleteCity(ctx context.Context, req api.DeleteCityRequestObject) (api.DeleteCityResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "city_write") {
		return api.DeleteCity403Response{}, nil
	}
	if err := h.store.DeleteCity(ctx, req.ID); err != nil {
		return nil, err
	}
	return api.DeleteCity204Response{}, nil
}

func (h *Handlers) GetCity(ctx context.Context, req api.GetCityRequestObject) (api.GetCityResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "city_read") {
		return api.GetCity404Response{}, nil
	}
	m, err := h.store.GetCity(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return api.GetCity404Response{}, nil
	}
	var resp api.City
	jsonConvert(m, &resp)
	return api.GetCity200JSONResponse(resp), nil
}

func (h *Handlers) UpdateCity(ctx context.Context, req api.UpdateCityRequestObject) (api.UpdateCityResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "city_write") {
		return api.UpdateCity403Response{}, nil
	}
	var m model.City
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.UpdateCity400Response{}, nil
	}
	m.ID = req.ID
	if err := h.store.UpdateCity(ctx, &m); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			return api.UpdateCity409Response{}, nil
		}
		return api.UpdateCity400Response{}, nil
	}
	var resp api.City
	jsonConvert(&m, &resp)
	return api.UpdateCity200JSONResponse(resp), nil
}

// ── Curricula ─────────────────────────────────────────────────────────────────

func (h *Handlers) ListCurriculums(ctx context.Context, req api.ListCurriculumsRequestObject) (api.ListCurriculumsResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "curriculum_read") {
		return api.ListCurriculums403Response{}, nil
	}
	list, err := h.store.ListCurricula(ctx)
	if err != nil {
		return nil, err
	}
	var resp []api.Curriculum
	jsonConvert(list, &resp)
	if resp == nil {
		resp = []api.Curriculum{}
	}
	return api.ListCurriculums200JSONResponse(resp), nil
}

func (h *Handlers) CreateCurriculum(ctx context.Context, req api.CreateCurriculumRequestObject) (api.CreateCurriculumResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "curriculum_write") {
		return api.CreateCurriculum403Response{}, nil
	}
	var m model.Curriculum
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.CreateCurriculum400Response{}, nil
	}
	if err := h.store.CreateCurriculum(ctx, &m); err != nil {
		return api.CreateCurriculum400Response{}, nil
	}
	var resp api.Curriculum
	jsonConvert(&m, &resp)
	return api.CreateCurriculum201JSONResponse(resp), nil
}

func (h *Handlers) DeleteCurriculum(ctx context.Context, req api.DeleteCurriculumRequestObject) (api.DeleteCurriculumResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "curriculum_write") {
		return api.DeleteCurriculum403Response{}, nil
	}
	if err := h.store.DeleteCurriculum(ctx, req.ID); err != nil {
		return nil, err
	}
	return api.DeleteCurriculum204Response{}, nil
}

func (h *Handlers) GetCurriculum(ctx context.Context, req api.GetCurriculumRequestObject) (api.GetCurriculumResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "curriculum_read") {
		return api.GetCurriculum404Response{}, nil
	}
	m, err := h.store.GetCurriculum(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return api.GetCurriculum404Response{}, nil
	}
	var resp api.Curriculum
	jsonConvert(m, &resp)
	return api.GetCurriculum200JSONResponse(resp), nil
}

func (h *Handlers) UpdateCurriculum(ctx context.Context, req api.UpdateCurriculumRequestObject) (api.UpdateCurriculumResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "curriculum_write") {
		return api.UpdateCurriculum403Response{}, nil
	}
	var m model.Curriculum
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.UpdateCurriculum400Response{}, nil
	}
	m.ID = req.ID
	if err := h.store.UpdateCurriculum(ctx, &m); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			return api.UpdateCurriculum409Response{}, nil
		}
		return api.UpdateCurriculum400Response{}, nil
	}
	var resp api.Curriculum
	jsonConvert(&m, &resp)
	return api.UpdateCurriculum200JSONResponse(resp), nil
}

// ── Exams ─────────────────────────────────────────────────────────────────────

func (h *Handlers) ListExams(ctx context.Context, req api.ListExamsRequestObject) (api.ListExamsResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "exam_read") {
		return api.ListExams403Response{}, nil
	}
	list, err := h.store.ListExams(ctx)
	if err != nil {
		return nil, err
	}
	var resp []api.Exam
	jsonConvert(list, &resp)
	if resp == nil {
		resp = []api.Exam{}
	}
	return api.ListExams200JSONResponse(resp), nil
}

func (h *Handlers) CreateExam(ctx context.Context, req api.CreateExamRequestObject) (api.CreateExamResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "exam_write_own") && !hasRole(claims, "exam_write_all") {
		return api.CreateExam403Response{}, nil
	}
	var m model.Exam
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.CreateExam400Response{}, nil
	}
	if err := h.store.CreateExam(ctx, &m); err != nil {
		return api.CreateExam400Response{}, nil
	}
	var resp api.Exam
	jsonConvert(&m, &resp)
	return api.CreateExam201JSONResponse(resp), nil
}

func (h *Handlers) DeleteExam(ctx context.Context, req api.DeleteExamRequestObject) (api.DeleteExamResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	isTeacher, err := h.store.IsTeacherOfExam(ctx, req.ID, claims.Sub)
	if err != nil {
		return nil, err
	}
	if !hasWriteAccess(claims, "exam", func() bool { return isTeacher }) {
		return api.DeleteExam403Response{}, nil
	}
	if err := h.store.DeleteExam(ctx, req.ID); err != nil {
		return nil, err
	}
	return api.DeleteExam204Response{}, nil
}

func (h *Handlers) GetExam(ctx context.Context, req api.GetExamRequestObject) (api.GetExamResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "exam_read") {
		return api.GetExam404Response{}, nil
	}
	m, err := h.store.GetExam(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return api.GetExam404Response{}, nil
	}
	var resp api.Exam
	jsonConvert(m, &resp)
	return api.GetExam200JSONResponse(resp), nil
}

func (h *Handlers) UpdateExam(ctx context.Context, req api.UpdateExamRequestObject) (api.UpdateExamResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	isTeacher, err := h.store.IsTeacherOfExam(ctx, req.ID, claims.Sub)
	if err != nil {
		return nil, err
	}
	if !hasWriteAccess(claims, "exam", func() bool { return isTeacher }) {
		return api.UpdateExam403Response{}, nil
	}
	var m model.Exam
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.UpdateExam400Response{}, nil
	}
	m.ID = req.ID
	if err := h.store.UpdateExam(ctx, &m); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			return api.UpdateExam409Response{}, nil
		}
		return api.UpdateExam400Response{}, nil
	}
	var resp api.Exam
	jsonConvert(&m, &resp)
	return api.UpdateExam200JSONResponse(resp), nil
}

// ── Grades ────────────────────────────────────────────────────────────────────

func (h *Handlers) ListGrades(ctx context.Context, req api.ListGradesRequestObject) (api.ListGradesResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "grade_read") {
		return api.ListGrades403Response{}, nil
	}
	list, err := h.store.ListGrades(ctx)
	if err != nil {
		return nil, err
	}
	var resp []api.Grade
	jsonConvert(list, &resp)
	if resp == nil {
		resp = []api.Grade{}
	}
	return api.ListGrades200JSONResponse(resp), nil
}

func (h *Handlers) CreateGrade(ctx context.Context, req api.CreateGradeRequestObject) (api.CreateGradeResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "grade_write_own") && !hasRole(claims, "grade_write_all") {
		return api.CreateGrade403Response{}, nil
	}
	var m model.Grade
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.CreateGrade400Response{}, nil
	}
	if err := h.store.CreateGrade(ctx, &m); err != nil {
		return api.CreateGrade400Response{}, nil
	}
	var resp api.Grade
	jsonConvert(&m, &resp)
	return api.CreateGrade201JSONResponse(resp), nil
}

func (h *Handlers) DeleteGrade(ctx context.Context, req api.DeleteGradeRequestObject) (api.DeleteGradeResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	isTeacher, err := h.store.IsTeacherOfGrade(ctx, req.ID, claims.Sub)
	if err != nil {
		return nil, err
	}
	if !hasWriteAccess(claims, "grade", func() bool { return isTeacher }) {
		return api.DeleteGrade403Response{}, nil
	}
	if err := h.store.DeleteGrade(ctx, req.ID); err != nil {
		return nil, err
	}
	return api.DeleteGrade204Response{}, nil
}

func (h *Handlers) GetGrade(ctx context.Context, req api.GetGradeRequestObject) (api.GetGradeResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "grade_read") {
		return api.GetGrade404Response{}, nil
	}
	m, err := h.store.GetGrade(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return api.GetGrade404Response{}, nil
	}
	var resp api.Grade
	jsonConvert(m, &resp)
	return api.GetGrade200JSONResponse(resp), nil
}

func (h *Handlers) UpdateGrade(ctx context.Context, req api.UpdateGradeRequestObject) (api.UpdateGradeResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	isTeacher, err := h.store.IsTeacherOfGrade(ctx, req.ID, claims.Sub)
	if err != nil {
		return nil, err
	}
	if !hasWriteAccess(claims, "grade", func() bool { return isTeacher }) {
		return api.UpdateGrade403Response{}, nil
	}
	var m model.Grade
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.UpdateGrade400Response{}, nil
	}
	m.ID = req.ID
	if err := h.store.UpdateGrade(ctx, &m); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			return api.UpdateGrade409Response{}, nil
		}
		return api.UpdateGrade400Response{}, nil
	}
	var resp api.Grade
	jsonConvert(&m, &resp)
	return api.UpdateGrade200JSONResponse(resp), nil
}

// ── Guardians ─────────────────────────────────────────────────────────────────

func (h *Handlers) ListGuardians(ctx context.Context, req api.ListGuardiansRequestObject) (api.ListGuardiansResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "guardian_read") {
		return api.ListGuardians403Response{}, nil
	}
	list, err := h.store.ListGuardians(ctx)
	if err != nil {
		return nil, err
	}
	var resp []api.Guardian
	jsonConvert(list, &resp)
	if resp == nil {
		resp = []api.Guardian{}
	}
	return api.ListGuardians200JSONResponse(resp), nil
}

func (h *Handlers) CreateGuardian(ctx context.Context, req api.CreateGuardianRequestObject) (api.CreateGuardianResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "guardian_write") {
		return api.CreateGuardian403Response{}, nil
	}
	var m model.Guardian
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.CreateGuardian400Response{}, nil
	}
	if err := h.store.CreateGuardian(ctx, &m); err != nil {
		return api.CreateGuardian400Response{}, nil
	}
	var resp api.Guardian
	jsonConvert(&m, &resp)
	return api.CreateGuardian201JSONResponse(resp), nil
}

func (h *Handlers) DeleteGuardian(ctx context.Context, req api.DeleteGuardianRequestObject) (api.DeleteGuardianResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "guardian_write") {
		return api.DeleteGuardian403Response{}, nil
	}
	if err := h.store.DeleteGuardian(ctx, req.ID); err != nil {
		return nil, err
	}
	return api.DeleteGuardian204Response{}, nil
}

func (h *Handlers) GetGuardian(ctx context.Context, req api.GetGuardianRequestObject) (api.GetGuardianResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "guardian_read") {
		return api.GetGuardian404Response{}, nil
	}
	m, err := h.store.GetGuardian(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return api.GetGuardian404Response{}, nil
	}
	var resp api.Guardian
	jsonConvert(m, &resp)
	return api.GetGuardian200JSONResponse(resp), nil
}

func (h *Handlers) UpdateGuardian(ctx context.Context, req api.UpdateGuardianRequestObject) (api.UpdateGuardianResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "guardian_write") {
		return api.UpdateGuardian403Response{}, nil
	}
	var m model.Guardian
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.UpdateGuardian400Response{}, nil
	}
	m.ID = req.ID
	if err := h.store.UpdateGuardian(ctx, &m); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			return api.UpdateGuardian409Response{}, nil
		}
		return api.UpdateGuardian400Response{}, nil
	}
	var resp api.Guardian
	jsonConvert(&m, &resp)
	return api.UpdateGuardian200JSONResponse(resp), nil
}

// ── Lessons ───────────────────────────────────────────────────────────────────

func (h *Handlers) ListLessons(ctx context.Context, req api.ListLessonsRequestObject) (api.ListLessonsResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "lesson_read") {
		return api.ListLessons403Response{}, nil
	}
	list, err := h.store.ListLessons(ctx)
	if err != nil {
		return nil, err
	}
	var resp []api.Lesson
	jsonConvert(list, &resp)
	if resp == nil {
		resp = []api.Lesson{}
	}
	return api.ListLessons200JSONResponse(resp), nil
}

func (h *Handlers) CreateLesson(ctx context.Context, req api.CreateLessonRequestObject) (api.CreateLessonResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "lesson_write_own") && !hasRole(claims, "lesson_write_all") {
		return api.CreateLesson403Response{}, nil
	}
	var m model.Lesson
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.CreateLesson400Response{}, nil
	}
	if err := h.store.CreateLesson(ctx, &m); err != nil {
		return api.CreateLesson400Response{}, nil
	}
	var resp api.Lesson
	jsonConvert(&m, &resp)
	return api.CreateLesson201JSONResponse(resp), nil
}

func (h *Handlers) DeleteLesson(ctx context.Context, req api.DeleteLessonRequestObject) (api.DeleteLessonResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	existing, err := h.store.GetLesson(ctx, req.ID)
	if err != nil || existing == nil {
		return api.DeleteLesson403Response{}, nil
	}
	if !hasWriteAccess(claims, "lesson", func() bool {
		return existing.Teacher != nil && existing.Teacher.Sub != nil && *existing.Teacher.Sub == claims.Sub
	}) {
		return api.DeleteLesson403Response{}, nil
	}
	if err := h.store.DeleteLesson(ctx, req.ID); err != nil {
		return nil, err
	}
	return api.DeleteLesson204Response{}, nil
}

func (h *Handlers) GetLesson(ctx context.Context, req api.GetLessonRequestObject) (api.GetLessonResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "lesson_read") {
		return api.GetLesson404Response{}, nil
	}
	m, err := h.store.GetLesson(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return api.GetLesson404Response{}, nil
	}
	var resp api.Lesson
	jsonConvert(m, &resp)
	return api.GetLesson200JSONResponse(resp), nil
}

func (h *Handlers) UpdateLesson(ctx context.Context, req api.UpdateLessonRequestObject) (api.UpdateLessonResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	existing, err := h.store.GetLesson(ctx, req.ID)
	if err != nil || existing == nil {
		return api.UpdateLesson400Response{}, nil
	}
	if !hasWriteAccess(claims, "lesson", func() bool {
		return existing.Teacher != nil && existing.Teacher.Sub != nil && *existing.Teacher.Sub == claims.Sub
	}) {
		return api.UpdateLesson403Response{}, nil
	}
	var m model.Lesson
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.UpdateLesson400Response{}, nil
	}
	m.ID = req.ID
	if err := h.store.UpdateLesson(ctx, &m); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			return api.UpdateLesson409Response{}, nil
		}
		return api.UpdateLesson400Response{}, nil
	}
	var resp api.Lesson
	jsonConvert(&m, &resp)
	return api.UpdateLesson200JSONResponse(resp), nil
}

// ── Locations ─────────────────────────────────────────────────────────────────

func (h *Handlers) ListLocations(ctx context.Context, req api.ListLocationsRequestObject) (api.ListLocationsResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "location_read") {
		return api.ListLocations403Response{}, nil
	}
	list, err := h.store.ListLocations(ctx)
	if err != nil {
		return nil, err
	}
	var resp []api.Location
	jsonConvert(list, &resp)
	if resp == nil {
		resp = []api.Location{}
	}
	return api.ListLocations200JSONResponse(resp), nil
}

func (h *Handlers) CreateLocation(ctx context.Context, req api.CreateLocationRequestObject) (api.CreateLocationResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "location_write") {
		return api.CreateLocation403Response{}, nil
	}
	var m model.Location
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.CreateLocation400Response{}, nil
	}
	if err := h.store.CreateLocation(ctx, &m); err != nil {
		return api.CreateLocation400Response{}, nil
	}
	var resp api.Location
	jsonConvert(&m, &resp)
	return api.CreateLocation201JSONResponse(resp), nil
}

func (h *Handlers) DeleteLocation(ctx context.Context, req api.DeleteLocationRequestObject) (api.DeleteLocationResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "location_write") {
		return api.DeleteLocation403Response{}, nil
	}
	if err := h.store.DeleteLocation(ctx, req.ID); err != nil {
		return nil, err
	}
	return api.DeleteLocation204Response{}, nil
}

func (h *Handlers) GetLocation(ctx context.Context, req api.GetLocationRequestObject) (api.GetLocationResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "location_read") {
		return api.GetLocation404Response{}, nil
	}
	m, err := h.store.GetLocation(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return api.GetLocation404Response{}, nil
	}
	var resp api.Location
	jsonConvert(m, &resp)
	return api.GetLocation200JSONResponse(resp), nil
}

func (h *Handlers) UpdateLocation(ctx context.Context, req api.UpdateLocationRequestObject) (api.UpdateLocationResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "location_write") {
		return api.UpdateLocation403Response{}, nil
	}
	var m model.Location
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.UpdateLocation400Response{}, nil
	}
	m.ID = req.ID
	if err := h.store.UpdateLocation(ctx, &m); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			return api.UpdateLocation409Response{}, nil
		}
		return api.UpdateLocation400Response{}, nil
	}
	var resp api.Location
	jsonConvert(&m, &resp)
	return api.UpdateLocation200JSONResponse(resp), nil
}

// ── Persons ───────────────────────────────────────────────────────────────────

func (h *Handlers) ListPersons(ctx context.Context, req api.ListPersonsRequestObject) (api.ListPersonsResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "person_read") {
		return api.ListPersons403Response{}, nil
	}
	list, err := h.store.ListPersons(ctx)
	if err != nil {
		return nil, err
	}
	var resp []api.Person
	jsonConvert(list, &resp)
	if resp == nil {
		resp = []api.Person{}
	}
	return api.ListPersons200JSONResponse(resp), nil
}

func (h *Handlers) CreatePerson(ctx context.Context, req api.CreatePersonRequestObject) (api.CreatePersonResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "person_write") {
		return api.CreatePerson403Response{}, nil
	}
	var m model.Person
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.CreatePerson400Response{}, nil
	}
	if err := h.store.CreatePerson(ctx, &m); err != nil {
		return api.CreatePerson400Response{}, nil
	}
	var resp api.Person
	jsonConvert(&m, &resp)
	return api.CreatePerson201JSONResponse(resp), nil
}

func (h *Handlers) DeletePerson(ctx context.Context, req api.DeletePersonRequestObject) (api.DeletePersonResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "person_write") {
		return api.DeletePerson403Response{}, nil
	}
	if err := h.store.DeletePerson(ctx, req.ID); err != nil {
		return nil, err
	}
	return api.DeletePerson204Response{}, nil
}

func (h *Handlers) GetPerson(ctx context.Context, req api.GetPersonRequestObject) (api.GetPersonResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "person_read") {
		return api.GetPerson404Response{}, nil
	}
	m, err := h.store.GetPerson(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return api.GetPerson404Response{}, nil
	}
	var resp api.Person
	jsonConvert(m, &resp)
	return api.GetPerson200JSONResponse(resp), nil
}

func (h *Handlers) UpdatePerson(ctx context.Context, req api.UpdatePersonRequestObject) (api.UpdatePersonResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "person_write") {
		return api.UpdatePerson403Response{}, nil
	}
	var m model.Person
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.UpdatePerson400Response{}, nil
	}
	m.ID = req.ID
	if err := h.store.UpdatePerson(ctx, &m); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			return api.UpdatePerson409Response{}, nil
		}
		return api.UpdatePerson400Response{}, nil
	}
	var resp api.Person
	jsonConvert(&m, &resp)
	return api.UpdatePerson200JSONResponse(resp), nil
}

// ── PostalCodes ───────────────────────────────────────────────────────────────

func (h *Handlers) ListPostalCodes(ctx context.Context, req api.ListPostalCodesRequestObject) (api.ListPostalCodesResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "postalcode_read") {
		return api.ListPostalCodes403Response{}, nil
	}
	list, err := h.store.ListPostalCodes(ctx)
	if err != nil {
		return nil, err
	}
	var resp []api.PostalCode
	jsonConvert(list, &resp)
	if resp == nil {
		resp = []api.PostalCode{}
	}
	return api.ListPostalCodes200JSONResponse(resp), nil
}

func (h *Handlers) CreatePostalCode(ctx context.Context, req api.CreatePostalCodeRequestObject) (api.CreatePostalCodeResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "postalcode_write") {
		return api.CreatePostalCode403Response{}, nil
	}
	var m model.PostalCode
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.CreatePostalCode400Response{}, nil
	}
	if err := h.store.CreatePostalCode(ctx, &m); err != nil {
		return api.CreatePostalCode400Response{}, nil
	}
	var resp api.PostalCode
	jsonConvert(&m, &resp)
	return api.CreatePostalCode201JSONResponse(resp), nil
}

func (h *Handlers) DeletePostalCode(ctx context.Context, req api.DeletePostalCodeRequestObject) (api.DeletePostalCodeResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "postalcode_write") {
		return api.DeletePostalCode403Response{}, nil
	}
	if err := h.store.DeletePostalCode(ctx, req.ID); err != nil {
		return nil, err
	}
	return api.DeletePostalCode204Response{}, nil
}

func (h *Handlers) GetPostalCode(ctx context.Context, req api.GetPostalCodeRequestObject) (api.GetPostalCodeResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "postalcode_read") {
		return api.GetPostalCode404Response{}, nil
	}
	m, err := h.store.GetPostalCode(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return api.GetPostalCode404Response{}, nil
	}
	var resp api.PostalCode
	jsonConvert(m, &resp)
	return api.GetPostalCode200JSONResponse(resp), nil
}

func (h *Handlers) UpdatePostalCode(ctx context.Context, req api.UpdatePostalCodeRequestObject) (api.UpdatePostalCodeResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "postalcode_write") {
		return api.UpdatePostalCode403Response{}, nil
	}
	var m model.PostalCode
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.UpdatePostalCode400Response{}, nil
	}
	m.ID = req.ID
	if err := h.store.UpdatePostalCode(ctx, &m); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			return api.UpdatePostalCode409Response{}, nil
		}
		return api.UpdatePostalCode400Response{}, nil
	}
	var resp api.PostalCode
	jsonConvert(&m, &resp)
	return api.UpdatePostalCode200JSONResponse(resp), nil
}

// ── Receipts ──────────────────────────────────────────────────────────────────

func (h *Handlers) ListReceipts(ctx context.Context, req api.ListReceiptsRequestObject) (api.ListReceiptsResponseObject, error) {
	// Auth is handled by global middleware for receipt routes.
	ownerID := ""
	if req.Params.Owner != nil {
		ownerID = *req.Params.Owner
	}
	list, err := h.store.List(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	var resp []api.Receipt
	jsonConvert(list, &resp)
	if resp == nil {
		resp = []api.Receipt{}
	}
	return api.ListReceipts200JSONResponse(resp), nil
}

func (h *Handlers) CreateReceipt(ctx context.Context, req api.CreateReceiptRequestObject) (api.CreateReceiptResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "leanschool_write") {
		return api.CreateReceipt400Response{}, nil
	}
	var m model.Receipt
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.CreateReceipt400Response{}, nil
	}
	if err := h.store.Create(ctx, &m); err != nil {
		return api.CreateReceipt400Response{}, nil
	}
	var resp api.Receipt
	jsonConvert(&m, &resp)
	return api.CreateReceipt201JSONResponse(resp), nil
}

func (h *Handlers) DeleteReceipt(ctx context.Context, req api.DeleteReceiptRequestObject) (api.DeleteReceiptResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "leanschool_write") {
		return api.DeleteReceipt404Response{}, nil
	}
	if err := h.store.Delete(ctx, req.ID); err != nil {
		return api.DeleteReceipt404Response{}, nil
	}
	return api.DeleteReceipt204Response{}, nil
}

func (h *Handlers) GetReceipt(ctx context.Context, req api.GetReceiptRequestObject) (api.GetReceiptResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "leanschool_read") {
		return api.GetReceipt404Response{}, nil
	}
	m, err := h.store.Get(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return api.GetReceipt404Response{}, nil
	}
	var resp api.Receipt
	jsonConvert(m, &resp)
	return api.GetReceipt200JSONResponse(resp), nil
}

func (h *Handlers) UpdateReceipt(ctx context.Context, req api.UpdateReceiptRequestObject) (api.UpdateReceiptResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "leanschool_write") {
		return api.UpdateReceipt404Response{}, nil
	}
	var m model.Receipt
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.UpdateReceipt404Response{}, nil
	}
	m.ID = req.ID
	if err := h.store.Update(ctx, &m); err != nil {
		return api.UpdateReceipt404Response{}, nil
	}
	var resp api.Receipt
	jsonConvert(&m, &resp)
	return api.UpdateReceipt200JSONResponse(resp), nil
}

// ── Rooms ─────────────────────────────────────────────────────────────────────

func (h *Handlers) ListRooms(ctx context.Context, req api.ListRoomsRequestObject) (api.ListRoomsResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "room_read") {
		return api.ListRooms403Response{}, nil
	}
	list, err := h.store.ListRooms(ctx)
	if err != nil {
		return nil, err
	}
	var resp []api.Room
	jsonConvert(list, &resp)
	if resp == nil {
		resp = []api.Room{}
	}
	return api.ListRooms200JSONResponse(resp), nil
}

func (h *Handlers) CreateRoom(ctx context.Context, req api.CreateRoomRequestObject) (api.CreateRoomResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "room_write") {
		return api.CreateRoom403Response{}, nil
	}
	var m model.Room
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.CreateRoom400Response{}, nil
	}
	if err := h.store.CreateRoom(ctx, &m); err != nil {
		return api.CreateRoom400Response{}, nil
	}
	var resp api.Room
	jsonConvert(&m, &resp)
	return api.CreateRoom201JSONResponse(resp), nil
}

func (h *Handlers) DeleteRoom(ctx context.Context, req api.DeleteRoomRequestObject) (api.DeleteRoomResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "room_write") {
		return api.DeleteRoom403Response{}, nil
	}
	if err := h.store.DeleteRoom(ctx, req.ID); err != nil {
		return nil, err
	}
	return api.DeleteRoom204Response{}, nil
}

func (h *Handlers) GetRoom(ctx context.Context, req api.GetRoomRequestObject) (api.GetRoomResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "room_read") {
		return api.GetRoom404Response{}, nil
	}
	m, err := h.store.GetRoom(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return api.GetRoom404Response{}, nil
	}
	var resp api.Room
	jsonConvert(m, &resp)
	return api.GetRoom200JSONResponse(resp), nil
}

func (h *Handlers) UpdateRoom(ctx context.Context, req api.UpdateRoomRequestObject) (api.UpdateRoomResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "room_write") {
		return api.UpdateRoom403Response{}, nil
	}
	var m model.Room
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.UpdateRoom400Response{}, nil
	}
	m.ID = req.ID
	if err := h.store.UpdateRoom(ctx, &m); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			return api.UpdateRoom409Response{}, nil
		}
		return api.UpdateRoom400Response{}, nil
	}
	var resp api.Room
	jsonConvert(&m, &resp)
	return api.UpdateRoom200JSONResponse(resp), nil
}

// ── SchoolClasses ─────────────────────────────────────────────────────────────

func (h *Handlers) ListSchoolClasss(ctx context.Context, req api.ListSchoolClasssRequestObject) (api.ListSchoolClasssResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "schoolclass_read") {
		return api.ListSchoolClasss403Response{}, nil
	}
	list, err := h.store.ListSchoolClasses(ctx)
	if err != nil {
		return nil, err
	}
	var resp []api.SchoolClass
	jsonConvert(list, &resp)
	if resp == nil {
		resp = []api.SchoolClass{}
	}
	return api.ListSchoolClasss200JSONResponse(resp), nil
}

func (h *Handlers) CreateSchoolClass(ctx context.Context, req api.CreateSchoolClassRequestObject) (api.CreateSchoolClassResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "schoolclass_write_all") {
		return api.CreateSchoolClass403Response{}, nil
	}
	var m model.SchoolClass
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.CreateSchoolClass400Response{}, nil
	}
	if err := h.store.CreateSchoolClass(ctx, &m); err != nil {
		return api.CreateSchoolClass400Response{}, nil
	}
	var resp api.SchoolClass
	jsonConvert(&m, &resp)
	return api.CreateSchoolClass201JSONResponse(resp), nil
}

func (h *Handlers) DeleteSchoolClass(ctx context.Context, req api.DeleteSchoolClassRequestObject) (api.DeleteSchoolClassResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	existing, err := h.store.GetSchoolClass(ctx, req.ID)
	if err != nil || existing == nil {
		return api.DeleteSchoolClass403Response{}, nil
	}
	if !hasWriteAccess(claims, "schoolclass", func() bool {
		for _, teacher := range existing.Teachers {
			if teacher.Sub != nil && *teacher.Sub == claims.Sub {
				return true
			}
		}
		return false
	}) {
		return api.DeleteSchoolClass403Response{}, nil
	}
	if err := h.store.DeleteSchoolClass(ctx, req.ID); err != nil {
		return nil, err
	}
	return api.DeleteSchoolClass204Response{}, nil
}

func (h *Handlers) GetSchoolClass(ctx context.Context, req api.GetSchoolClassRequestObject) (api.GetSchoolClassResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "schoolclass_read") {
		return api.GetSchoolClass404Response{}, nil
	}
	m, err := h.store.GetSchoolClass(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return api.GetSchoolClass404Response{}, nil
	}
	var resp api.SchoolClass
	jsonConvert(m, &resp)
	return api.GetSchoolClass200JSONResponse(resp), nil
}

func (h *Handlers) UpdateSchoolClass(ctx context.Context, req api.UpdateSchoolClassRequestObject) (api.UpdateSchoolClassResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	existing, err := h.store.GetSchoolClass(ctx, req.ID)
	if err != nil || existing == nil {
		return api.UpdateSchoolClass400Response{}, nil
	}
	if !hasWriteAccess(claims, "schoolclass", func() bool {
		for _, teacher := range existing.Teachers {
			if teacher.Sub != nil && *teacher.Sub == claims.Sub {
				return true
			}
		}
		return false
	}) {
		return api.UpdateSchoolClass403Response{}, nil
	}
	var m model.SchoolClass
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.UpdateSchoolClass400Response{}, nil
	}
	m.ID = req.ID
	if err := h.store.UpdateSchoolClass(ctx, &m); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			return api.UpdateSchoolClass409Response{}, nil
		}
		return api.UpdateSchoolClass400Response{}, nil
	}
	var resp api.SchoolClass
	jsonConvert(&m, &resp)
	return api.UpdateSchoolClass200JSONResponse(resp), nil
}

// ── SchoolYears ───────────────────────────────────────────────────────────────

func (h *Handlers) ListSchoolYears(ctx context.Context, req api.ListSchoolYearsRequestObject) (api.ListSchoolYearsResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "schoolyear_read") {
		return api.ListSchoolYears403Response{}, nil
	}
	list, err := h.store.ListSchoolYears(ctx)
	if err != nil {
		return nil, err
	}
	var resp []api.SchoolYear
	jsonConvert(list, &resp)
	if resp == nil {
		resp = []api.SchoolYear{}
	}
	return api.ListSchoolYears200JSONResponse(resp), nil
}

func (h *Handlers) CreateSchoolYear(ctx context.Context, req api.CreateSchoolYearRequestObject) (api.CreateSchoolYearResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "schoolyear_write") {
		return api.CreateSchoolYear403Response{}, nil
	}
	var m model.SchoolYear
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.CreateSchoolYear400Response{}, nil
	}
	if err := h.store.CreateSchoolYear(ctx, &m); err != nil {
		return api.CreateSchoolYear400Response{}, nil
	}
	var resp api.SchoolYear
	jsonConvert(&m, &resp)
	return api.CreateSchoolYear201JSONResponse(resp), nil
}

func (h *Handlers) DeleteSchoolYear(ctx context.Context, req api.DeleteSchoolYearRequestObject) (api.DeleteSchoolYearResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "schoolyear_write") {
		return api.DeleteSchoolYear403Response{}, nil
	}
	if err := h.store.DeleteSchoolYear(ctx, req.ID); err != nil {
		return nil, err
	}
	return api.DeleteSchoolYear204Response{}, nil
}

func (h *Handlers) GetSchoolYear(ctx context.Context, req api.GetSchoolYearRequestObject) (api.GetSchoolYearResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "schoolyear_read") {
		return api.GetSchoolYear404Response{}, nil
	}
	m, err := h.store.GetSchoolYear(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return api.GetSchoolYear404Response{}, nil
	}
	var resp api.SchoolYear
	jsonConvert(m, &resp)
	return api.GetSchoolYear200JSONResponse(resp), nil
}

func (h *Handlers) UpdateSchoolYear(ctx context.Context, req api.UpdateSchoolYearRequestObject) (api.UpdateSchoolYearResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "schoolyear_write") {
		return api.UpdateSchoolYear403Response{}, nil
	}
	var m model.SchoolYear
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.UpdateSchoolYear400Response{}, nil
	}
	m.ID = req.ID
	if err := h.store.UpdateSchoolYear(ctx, &m); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			return api.UpdateSchoolYear409Response{}, nil
		}
		return api.UpdateSchoolYear400Response{}, nil
	}
	var resp api.SchoolYear
	jsonConvert(&m, &resp)
	return api.UpdateSchoolYear200JSONResponse(resp), nil
}

// ── Students ──────────────────────────────────────────────────────────────────

func (h *Handlers) ListStudents(ctx context.Context, req api.ListStudentsRequestObject) (api.ListStudentsResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "student_read") {
		return api.ListStudents403Response{}, nil
	}
	list, err := h.store.ListStudents(ctx)
	if err != nil {
		return nil, err
	}
	var resp []api.Student
	jsonConvert(list, &resp)
	if resp == nil {
		resp = []api.Student{}
	}
	return api.ListStudents200JSONResponse(resp), nil
}

func (h *Handlers) CreateStudent(ctx context.Context, req api.CreateStudentRequestObject) (api.CreateStudentResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "student_write") {
		return api.CreateStudent403Response{}, nil
	}
	var m model.Student
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.CreateStudent400Response{}, nil
	}
	if err := h.store.CreateStudent(ctx, &m); err != nil {
		return api.CreateStudent400Response{}, nil
	}
	var resp api.Student
	jsonConvert(&m, &resp)
	return api.CreateStudent201JSONResponse(resp), nil
}

func (h *Handlers) DeleteStudent(ctx context.Context, req api.DeleteStudentRequestObject) (api.DeleteStudentResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "student_write") {
		return api.DeleteStudent403Response{}, nil
	}
	if err := h.store.DeleteStudent(ctx, req.ID); err != nil {
		return nil, err
	}
	return api.DeleteStudent204Response{}, nil
}

func (h *Handlers) GetStudent(ctx context.Context, req api.GetStudentRequestObject) (api.GetStudentResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "student_read") {
		return api.GetStudent404Response{}, nil
	}
	m, err := h.store.GetStudent(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return api.GetStudent404Response{}, nil
	}
	var resp api.Student
	jsonConvert(m, &resp)
	return api.GetStudent200JSONResponse(resp), nil
}

func (h *Handlers) UpdateStudent(ctx context.Context, req api.UpdateStudentRequestObject) (api.UpdateStudentResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "student_write") {
		return api.UpdateStudent403Response{}, nil
	}
	var m model.Student
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.UpdateStudent400Response{}, nil
	}
	m.ID = req.ID
	if err := h.store.UpdateStudent(ctx, &m); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			return api.UpdateStudent409Response{}, nil
		}
		return api.UpdateStudent400Response{}, nil
	}
	var resp api.Student
	jsonConvert(&m, &resp)
	return api.UpdateStudent200JSONResponse(resp), nil
}

// ── Subjects ──────────────────────────────────────────────────────────────────

func (h *Handlers) ListSubjects(ctx context.Context, req api.ListSubjectsRequestObject) (api.ListSubjectsResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "subject_read") {
		return api.ListSubjects403Response{}, nil
	}
	list, err := h.store.ListSubjects(ctx)
	if err != nil {
		return nil, err
	}
	var resp []api.Subject
	jsonConvert(list, &resp)
	if resp == nil {
		resp = []api.Subject{}
	}
	return api.ListSubjects200JSONResponse(resp), nil
}

func (h *Handlers) CreateSubject(ctx context.Context, req api.CreateSubjectRequestObject) (api.CreateSubjectResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "subject_write") {
		return api.CreateSubject403Response{}, nil
	}
	var m model.Subject
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.CreateSubject400Response{}, nil
	}
	if err := h.store.CreateSubject(ctx, &m); err != nil {
		return api.CreateSubject400Response{}, nil
	}
	var resp api.Subject
	jsonConvert(&m, &resp)
	return api.CreateSubject201JSONResponse(resp), nil
}

func (h *Handlers) DeleteSubject(ctx context.Context, req api.DeleteSubjectRequestObject) (api.DeleteSubjectResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "subject_write") {
		return api.DeleteSubject403Response{}, nil
	}
	if err := h.store.DeleteSubject(ctx, req.ID); err != nil {
		return nil, err
	}
	return api.DeleteSubject204Response{}, nil
}

func (h *Handlers) GetSubject(ctx context.Context, req api.GetSubjectRequestObject) (api.GetSubjectResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "subject_read") {
		return api.GetSubject404Response{}, nil
	}
	m, err := h.store.GetSubject(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return api.GetSubject404Response{}, nil
	}
	var resp api.Subject
	jsonConvert(m, &resp)
	return api.GetSubject200JSONResponse(resp), nil
}

func (h *Handlers) UpdateSubject(ctx context.Context, req api.UpdateSubjectRequestObject) (api.UpdateSubjectResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "subject_write") {
		return api.UpdateSubject403Response{}, nil
	}
	var m model.Subject
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.UpdateSubject400Response{}, nil
	}
	m.ID = req.ID
	if err := h.store.UpdateSubject(ctx, &m); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			return api.UpdateSubject409Response{}, nil
		}
		return api.UpdateSubject400Response{}, nil
	}
	var resp api.Subject
	jsonConvert(&m, &resp)
	return api.UpdateSubject200JSONResponse(resp), nil
}

// ── Teachers ──────────────────────────────────────────────────────────────────

func (h *Handlers) ListTeachers(ctx context.Context, req api.ListTeachersRequestObject) (api.ListTeachersResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "teacher_read") {
		return api.ListTeachers403Response{}, nil
	}
	list, err := h.store.ListTeachers(ctx)
	if err != nil {
		return nil, err
	}
	var resp []api.Teacher
	jsonConvert(list, &resp)
	if resp == nil {
		resp = []api.Teacher{}
	}
	return api.ListTeachers200JSONResponse(resp), nil
}

func (h *Handlers) CreateTeacher(ctx context.Context, req api.CreateTeacherRequestObject) (api.CreateTeacherResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "teacher_write_all") {
		return api.CreateTeacher403Response{}, nil
	}
	var m model.Teacher
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.CreateTeacher400Response{}, nil
	}
	if err := h.store.CreateTeacher(ctx, &m); err != nil {
		return api.CreateTeacher400Response{}, nil
	}
	var resp api.Teacher
	jsonConvert(&m, &resp)
	return api.CreateTeacher201JSONResponse(resp), nil
}

func (h *Handlers) DeleteTeacher(ctx context.Context, req api.DeleteTeacherRequestObject) (api.DeleteTeacherResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	existing, err := h.store.GetTeacher(ctx, req.ID)
	if err != nil || existing == nil {
		return api.DeleteTeacher403Response{}, nil
	}
	if !hasWriteAccess(claims, "teacher", func() bool {
		return existing.Sub != nil && claims.Sub == *existing.Sub
	}) {
		return api.DeleteTeacher403Response{}, nil
	}
	if err := h.store.DeleteTeacher(ctx, req.ID); err != nil {
		return nil, err
	}
	return api.DeleteTeacher204Response{}, nil
}

func (h *Handlers) GetTeacher(ctx context.Context, req api.GetTeacherRequestObject) (api.GetTeacherResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	if !hasRole(claims, "teacher_read") {
		return api.GetTeacher404Response{}, nil
	}
	m, err := h.store.GetTeacher(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return api.GetTeacher404Response{}, nil
	}
	var resp api.Teacher
	jsonConvert(m, &resp)
	return api.GetTeacher200JSONResponse(resp), nil
}

func (h *Handlers) UpdateTeacher(ctx context.Context, req api.UpdateTeacherRequestObject) (api.UpdateTeacherResponseObject, error) {
	claims := ClaimsFromContext(ctx)
	existing, err := h.store.GetTeacher(ctx, req.ID)
	if err != nil || existing == nil {
		return api.UpdateTeacher400Response{}, nil
	}
	if !hasWriteAccess(claims, "teacher", func() bool {
		return existing.Sub != nil && claims.Sub == *existing.Sub
	}) {
		return api.UpdateTeacher403Response{}, nil
	}
	var m model.Teacher
	if err := jsonConvert(req.Body, &m); err != nil {
		return api.UpdateTeacher400Response{}, nil
	}
	m.ID = req.ID
	if err := h.store.UpdateTeacher(ctx, &m); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			return api.UpdateTeacher409Response{}, nil
		}
		return api.UpdateTeacher400Response{}, nil
	}
	var resp api.Teacher
	jsonConvert(&m, &resp)
	return api.UpdateTeacher200JSONResponse(resp), nil
}
