package storage

import (
	"context"
	"database/sql"
	"fmt"

	model "github.com/Joel-Haeberli/leanschool-model"
)

// ── SchoolYear ────────────────────────────────────────────────────────────────

func (p *Postgres) CreateSchoolYear(ctx context.Context, sy *model.SchoolYear) error {
	sy.ID = newID()
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO school_years (id, from_dt, to_dt, version) VALUES ($1, $2, $3, 0)`,
		sy.ID, nullableTime(sy.From), nullableTime(sy.To),
	)
	return err
}

func (p *Postgres) GetSchoolYear(ctx context.Context, id string) (*model.SchoolYear, error) {
	sy := &model.SchoolYear{}
	var from, to sql.NullTime
	err := p.db.QueryRowContext(ctx,
		`SELECT id, from_dt, to_dt, version FROM school_years WHERE id = $1`, id,
	).Scan(&sy.ID, &from, &to, &sy.Version)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying school year: %w", err)
	}
	if from.Valid {
		t := from.Time
		sy.From = &t
	}
	if to.Valid {
		t := to.Time
		sy.To = &t
	}
	return sy, nil
}

func (p *Postgres) ListSchoolYears(ctx context.Context) ([]*model.SchoolYear, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT id, from_dt, to_dt, version FROM school_years ORDER BY from_dt DESC`)
	if err != nil {
		return nil, fmt.Errorf("listing school years: %w", err)
	}
	defer rows.Close()
	var years []*model.SchoolYear
	for rows.Next() {
		sy := &model.SchoolYear{}
		var from, to sql.NullTime
		if err := rows.Scan(&sy.ID, &from, &to, &sy.Version); err != nil {
			return nil, err
		}
		if from.Valid {
			t := from.Time
			sy.From = &t
		}
		if to.Valid {
			t := to.Time
			sy.To = &t
		}
		years = append(years, sy)
	}
	return years, rows.Err()
}

func (p *Postgres) UpdateSchoolYear(ctx context.Context, sy *model.SchoolYear) error {
	res, err := p.db.ExecContext(ctx,
		`UPDATE school_years SET from_dt=$1, to_dt=$2, version=version+1 WHERE id=$3 AND version=$4`,
		nullableTime(sy.From), nullableTime(sy.To), sy.ID, sy.Version,
	)
	if err != nil {
		return fmt.Errorf("updating school year: %w", err)
	}
	return optimisticCheck(ctx, p.db, res, "school_years", sy.ID, &sy.Version)
}

func (p *Postgres) DeleteSchoolYear(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "school_years", id)
}

// ── Curriculum ────────────────────────────────────────────────────────────────

func (p *Postgres) CreateCurriculum(ctx context.Context, c *model.Curriculum) error {
	c.ID = newID()
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO curricula (id, name, active_since, active_until, active_from, version)
		 VALUES ($1, $2, $3, $4, $5, 0)`,
		c.ID, nullableString(c.Name),
		nullableTime(c.ActiveSince), nullableTime(c.ActiveUntil), nullableTime(c.ActiveFrom),
	)
	return err
}

func (p *Postgres) GetCurriculum(ctx context.Context, id string) (*model.Curriculum, error) {
	c := &model.Curriculum{}
	var since, until, from sql.NullTime
	err := p.db.QueryRowContext(ctx,
		`SELECT id, COALESCE(name,''), active_since, active_until, active_from, version
		 FROM curricula WHERE id = $1`, id,
	).Scan(&c.ID, &c.Name, &since, &until, &from, &c.Version)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying curriculum: %w", err)
	}
	if since.Valid {
		t := since.Time
		c.ActiveSince = &t
	}
	if until.Valid {
		t := until.Time
		c.ActiveUntil = &t
	}
	if from.Valid {
		t := from.Time
		c.ActiveFrom = &t
	}
	return c, nil
}

func (p *Postgres) ListCurricula(ctx context.Context) ([]*model.Curriculum, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT id, COALESCE(name,''), active_since, active_until, active_from, version
		 FROM curricula ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("listing curricula: %w", err)
	}
	defer rows.Close()
	var curricula []*model.Curriculum
	for rows.Next() {
		c := &model.Curriculum{}
		var since, until, from sql.NullTime
		if err := rows.Scan(&c.ID, &c.Name, &since, &until, &from, &c.Version); err != nil {
			return nil, err
		}
		if since.Valid {
			t := since.Time
			c.ActiveSince = &t
		}
		if until.Valid {
			t := until.Time
			c.ActiveUntil = &t
		}
		if from.Valid {
			t := from.Time
			c.ActiveFrom = &t
		}
		curricula = append(curricula, c)
	}
	return curricula, rows.Err()
}

func (p *Postgres) UpdateCurriculum(ctx context.Context, c *model.Curriculum) error {
	res, err := p.db.ExecContext(ctx,
		`UPDATE curricula SET name=$1, active_since=$2, active_until=$3, active_from=$4, version=version+1
		 WHERE id=$5 AND version=$6`,
		nullableString(c.Name),
		nullableTime(c.ActiveSince), nullableTime(c.ActiveUntil), nullableTime(c.ActiveFrom),
		c.ID, c.Version,
	)
	if err != nil {
		return fmt.Errorf("updating curriculum: %w", err)
	}
	return optimisticCheck(ctx, p.db, res, "curricula", c.ID, &c.Version)
}

func (p *Postgres) DeleteCurriculum(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "curricula", id)
}

// ── Subject ───────────────────────────────────────────────────────────────────

func (p *Postgres) CreateSubject(ctx context.Context, s *model.Subject) error {
	s.ID = newID()
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO subjects (id, name, version) VALUES ($1, $2, 0)`,
		s.ID, nullableString(s.Name),
	); err != nil {
		return fmt.Errorf("inserting subject: %w", err)
	}
	if err := syncSubjectTeachers(ctx, tx, s.ID, s.Teachers); err != nil {
		return err
	}
	return tx.Commit()
}

func (p *Postgres) GetSubject(ctx context.Context, id string) (*model.Subject, error) {
	s := &model.Subject{}
	err := p.db.QueryRowContext(ctx,
		`SELECT id, COALESCE(name,''), version FROM subjects WHERE id = $1`, id,
	).Scan(&s.ID, &s.Name, &s.Version)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying subject: %w", err)
	}
	teachers, err := loadSubjectTeachers(ctx, p.db, id)
	if err != nil {
		return nil, err
	}
	s.Teachers = teachers
	return s, nil
}

func (p *Postgres) ListSubjects(ctx context.Context) ([]*model.Subject, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT id, COALESCE(name,''), version FROM subjects ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("listing subjects: %w", err)
	}
	defer rows.Close()
	var subjects []*model.Subject
	for rows.Next() {
		s := &model.Subject{}
		if err := rows.Scan(&s.ID, &s.Name, &s.Version); err != nil {
			return nil, err
		}
		s.Teachers = []model.Teacher{}
		subjects = append(subjects, s)
	}
	return subjects, rows.Err()
}

func (p *Postgres) UpdateSubject(ctx context.Context, s *model.Subject) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx,
		`UPDATE subjects SET name=$1, version=version+1 WHERE id=$2 AND version=$3`,
		nullableString(s.Name), s.ID, s.Version,
	)
	if err != nil {
		return fmt.Errorf("updating subject: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		var exists bool
		_ = p.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM subjects WHERE id=$1)`, s.ID).Scan(&exists)
		if !exists {
			return fmt.Errorf("subject not found")
		}
		return ErrOptimisticLock
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM subject_teachers WHERE subject_id=$1`, s.ID); err != nil {
		return fmt.Errorf("clearing subject teachers: %w", err)
	}
	if err := syncSubjectTeachers(ctx, tx, s.ID, s.Teachers); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.Version++
	return nil
}

func (p *Postgres) DeleteSubject(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "subjects", id)
}

func syncSubjectTeachers(ctx context.Context, tx *sql.Tx, subjectID string, teachers []model.Teacher) error {
	for _, t := range teachers {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO subject_teachers (subject_id, teacher_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			subjectID, t.ID,
		); err != nil {
			return fmt.Errorf("linking teacher to subject: %w", err)
		}
	}
	return nil
}

func loadSubjectTeachers(ctx context.Context, db *sql.DB, subjectID string) ([]model.Teacher, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT p.id, p.name, p.prename, p.version
		FROM persons p
		JOIN teachers t ON t.person_id = p.id
		JOIN subject_teachers st ON st.teacher_id = p.id
		WHERE st.subject_id = $1
		ORDER BY p.name`, subjectID,
	)
	if err != nil {
		return nil, fmt.Errorf("loading subject teachers: %w", err)
	}
	defer rows.Close()
	var teachers []model.Teacher
	for rows.Next() {
		t := model.Teacher{}
		if err := rows.Scan(&t.ID, &t.Name, &t.Prename, &t.Version); err != nil {
			return nil, err
		}
		teachers = append(teachers, t)
	}
	if teachers == nil {
		teachers = []model.Teacher{}
	}
	return teachers, rows.Err()
}

// ── Lesson ────────────────────────────────────────────────────────────────────

func (p *Postgres) CreateLesson(ctx context.Context, l *model.Lesson) error {
	l.ID = newID()
	teacherID, classID, subjectID, roomID := "", "", "", ""
	if l.Teacher != nil {
		teacherID = l.Teacher.ID
	}
	if l.SchoolClass != nil {
		classID = l.SchoolClass.ID
	}
	if l.Subject != nil {
		subjectID = l.Subject.ID
	}
	if l.Room != nil {
		roomID = l.Room.ID
	}
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO lessons (id, teacher_id, school_class_id, subject_id, room_id, day_of_week, period, start_time, end_time, version)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 0)`,
		l.ID, nullableString(teacherID), nullableString(classID), nullableString(subjectID), nullableString(roomID),
		nullableInt(l.DayOfWeek), nullableInt(l.Period), nullableStringPtr(l.StartTime), nullableStringPtr(l.EndTime),
	)
	return err
}

func (p *Postgres) GetLesson(ctx context.Context, id string) (*model.Lesson, error) {
	l := &model.Lesson{}
	var tID, cID, sID sql.NullString
	var tName, tPrename, cName, sName sql.NullString
	var tVer, cVer, sVer sql.NullInt64
	var dayOfWeek, period sql.NullInt64
	var startTime, endTime sql.NullString
	var rmID, rmName, rmRoomType sql.NullString
	var rmVer sql.NullInt64
	err := p.db.QueryRowContext(ctx, `
		SELECT l.id, l.version, l.day_of_week, l.period, l.start_time, l.end_time,
		       t.person_id, tp.name, tp.prename, tp.version,
		       sc.id, sc.name, sc.version,
		       sub.id, sub.name, sub.version,
		       rm.id, rm.name, rm.room_type, rm.version
		FROM lessons l
		LEFT JOIN teachers t ON t.person_id = l.teacher_id
		LEFT JOIN persons tp ON tp.id = t.person_id
		LEFT JOIN school_classes sc ON sc.id = l.school_class_id
		LEFT JOIN subjects sub ON sub.id = l.subject_id
		LEFT JOIN rooms rm ON rm.id = l.room_id
		WHERE l.id = $1`, id,
	).Scan(
		&l.ID, &l.Version, &dayOfWeek, &period, &startTime, &endTime,
		&tID, &tName, &tPrename, &tVer,
		&cID, &cName, &cVer,
		&sID, &sName, &sVer,
		&rmID, &rmName, &rmRoomType, &rmVer,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying lesson: %w", err)
	}
	if dayOfWeek.Valid {
		v := int(dayOfWeek.Int64)
		l.DayOfWeek = &v
	}
	if period.Valid {
		v := int(period.Int64)
		l.Period = &v
	}
	if startTime.Valid {
		l.StartTime = &startTime.String
	}
	if endTime.Valid {
		l.EndTime = &endTime.String
	}
	if tID.Valid {
		t := &model.Teacher{}
		t.ID = tID.String
		t.Name = tName.String
		t.Prename = tPrename.String
		t.Version = int(tVer.Int64)
		l.Teacher = t
	}
	if cID.Valid {
		l.SchoolClass = &model.SchoolClass{ID: cID.String, Name: cName.String, Version: int(cVer.Int64), Teachers: []model.Teacher{}, Students: []model.Student{}}
	}
	if sID.Valid {
		l.Subject = &model.Subject{ID: sID.String, Name: sName.String, Version: int(sVer.Int64), Teachers: []model.Teacher{}}
	}
	if rmID.Valid {
		rm := &model.Room{ID: rmID.String, Name: rmName.String, Version: int(rmVer.Int64)}
		if rmRoomType.Valid {
			rt := model.RoomType(rmRoomType.String)
		rm.RoomType = &rt
		}
		l.Room = rm
	}
	return l, nil
}

func (p *Postgres) ListLessons(ctx context.Context) ([]*model.Lesson, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT l.id, l.version, l.day_of_week, l.period, l.start_time, l.end_time,
		       t.person_id, tp.name, tp.prename, tp.version,
		       sc.id, sc.name, sc.version,
		       sub.id, sub.name, sub.version,
		       rm.id, rm.name, rm.room_type, rm.version
		FROM lessons l
		LEFT JOIN teachers t ON t.person_id = l.teacher_id
		LEFT JOIN persons tp ON tp.id = t.person_id
		LEFT JOIN school_classes sc ON sc.id = l.school_class_id
		LEFT JOIN subjects sub ON sub.id = l.subject_id
		LEFT JOIN rooms rm ON rm.id = l.room_id
		ORDER BY l.id`)
	if err != nil {
		return nil, fmt.Errorf("listing lessons: %w", err)
	}
	defer rows.Close()
	var lessons []*model.Lesson
	for rows.Next() {
		l := &model.Lesson{}
		var tID, cID, sID sql.NullString
		var tName, tPrename, cName, sName sql.NullString
		var tVer, cVer, sVer sql.NullInt64
		var dayOfWeek, periodVal sql.NullInt64
		var startTime, endTime sql.NullString
		var rmID, rmName, rmRoomType sql.NullString
		var rmVer sql.NullInt64
		if err := rows.Scan(
			&l.ID, &l.Version, &dayOfWeek, &periodVal, &startTime, &endTime,
			&tID, &tName, &tPrename, &tVer,
			&cID, &cName, &cVer,
			&sID, &sName, &sVer,
			&rmID, &rmName, &rmRoomType, &rmVer,
		); err != nil {
			return nil, err
		}
		if dayOfWeek.Valid {
			v := int(dayOfWeek.Int64)
			l.DayOfWeek = &v
		}
		if periodVal.Valid {
			v := int(periodVal.Int64)
			l.Period = &v
		}
		if startTime.Valid {
			l.StartTime = &startTime.String
		}
		if endTime.Valid {
			l.EndTime = &endTime.String
		}
		if tID.Valid {
			t := &model.Teacher{}
			t.ID = tID.String
			t.Name = tName.String
			t.Prename = tPrename.String
			t.Version = int(tVer.Int64)
			l.Teacher = t
		}
		if cID.Valid {
			l.SchoolClass = &model.SchoolClass{ID: cID.String, Name: cName.String, Version: int(cVer.Int64), Teachers: []model.Teacher{}, Students: []model.Student{}}
		}
		if sID.Valid {
			l.Subject = &model.Subject{ID: sID.String, Name: sName.String, Version: int(sVer.Int64), Teachers: []model.Teacher{}}
		}
		if rmID.Valid {
			rm := &model.Room{ID: rmID.String, Name: rmName.String, Version: int(rmVer.Int64)}
			if rmRoomType.Valid {
				rt := model.RoomType(rmRoomType.String)
		rm.RoomType = &rt
			}
			l.Room = rm
		}
		lessons = append(lessons, l)
	}
	return lessons, rows.Err()
}

func (p *Postgres) UpdateLesson(ctx context.Context, l *model.Lesson) error {
	teacherID, classID, subjectID, roomID := "", "", "", ""
	if l.Teacher != nil {
		teacherID = l.Teacher.ID
	}
	if l.SchoolClass != nil {
		classID = l.SchoolClass.ID
	}
	if l.Subject != nil {
		subjectID = l.Subject.ID
	}
	if l.Room != nil {
		roomID = l.Room.ID
	}
	res, err := p.db.ExecContext(ctx,
		`UPDATE lessons SET teacher_id=$1, school_class_id=$2, subject_id=$3, room_id=$4,
		 day_of_week=$5, period=$6, start_time=$7, end_time=$8, version=version+1
		 WHERE id=$9 AND version=$10`,
		nullableString(teacherID), nullableString(classID), nullableString(subjectID), nullableString(roomID),
		nullableInt(l.DayOfWeek), nullableInt(l.Period), nullableStringPtr(l.StartTime), nullableStringPtr(l.EndTime),
		l.ID, l.Version,
	)
	if err != nil {
		return fmt.Errorf("updating lesson: %w", err)
	}
	return optimisticCheck(ctx, p.db, res, "lessons", l.ID, &l.Version)
}

func (p *Postgres) DeleteLesson(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "lessons", id)
}

// ── Exam ──────────────────────────────────────────────────────────────────────

func (p *Postgres) CreateExam(ctx context.Context, e *model.Exam) error {
	e.ID = newID()
	classID := ""
	if e.SchoolClass != nil {
		classID = e.SchoolClass.ID
	}
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO exams (id, school_class_id, version) VALUES ($1, $2, 0)`,
		e.ID, nullableString(classID),
	)
	return err
}

func (p *Postgres) GetExam(ctx context.Context, id string) (*model.Exam, error) {
	e := &model.Exam{}
	var cID, cName sql.NullString
	var cVer sql.NullInt64
	err := p.db.QueryRowContext(ctx, `
		SELECT e.id, e.version, sc.id, sc.name, sc.version
		FROM exams e
		LEFT JOIN school_classes sc ON sc.id = e.school_class_id
		WHERE e.id = $1`, id,
	).Scan(&e.ID, &e.Version, &cID, &cName, &cVer)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying exam: %w", err)
	}
	if cID.Valid {
		e.SchoolClass = &model.SchoolClass{ID: cID.String, Name: cName.String, Version: int(cVer.Int64), Teachers: []model.Teacher{}, Students: []model.Student{}}
	}
	return e, nil
}

func (p *Postgres) ListExams(ctx context.Context) ([]*model.Exam, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT e.id, e.version, sc.id, sc.name, sc.version
		FROM exams e
		LEFT JOIN school_classes sc ON sc.id = e.school_class_id
		ORDER BY e.id`)
	if err != nil {
		return nil, fmt.Errorf("listing exams: %w", err)
	}
	defer rows.Close()
	var exams []*model.Exam
	for rows.Next() {
		e := &model.Exam{}
		var cID, cName sql.NullString
		var cVer sql.NullInt64
		if err := rows.Scan(&e.ID, &e.Version, &cID, &cName, &cVer); err != nil {
			return nil, err
		}
		if cID.Valid {
			e.SchoolClass = &model.SchoolClass{ID: cID.String, Name: cName.String, Version: int(cVer.Int64), Teachers: []model.Teacher{}, Students: []model.Student{}}
		}
		exams = append(exams, e)
	}
	return exams, rows.Err()
}

func (p *Postgres) UpdateExam(ctx context.Context, e *model.Exam) error {
	classID := ""
	if e.SchoolClass != nil {
		classID = e.SchoolClass.ID
	}
	res, err := p.db.ExecContext(ctx,
		`UPDATE exams SET school_class_id=$1, version=version+1 WHERE id=$2 AND version=$3`,
		nullableString(classID), e.ID, e.Version,
	)
	if err != nil {
		return fmt.Errorf("updating exam: %w", err)
	}
	return optimisticCheck(ctx, p.db, res, "exams", e.ID, &e.Version)
}

func (p *Postgres) DeleteExam(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "exams", id)
}

// ── Grade ─────────────────────────────────────────────────────────────────────

func (p *Postgres) CreateGrade(ctx context.Context, g *model.Grade) error {
	g.ID = newID()
	examID, studentID := "", ""
	if g.Exam != nil {
		examID = g.Exam.ID
	}
	if g.Student != nil {
		studentID = g.Student.ID
	}
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO grades (id, grade, exam_id, student_id, version) VALUES ($1, $2, $3, $4, 0)`,
		g.ID, nullableFloat64(g.Grade), nullableString(examID), nullableString(studentID),
	)
	return err
}

func (p *Postgres) GetGrade(ctx context.Context, id string) (*model.Grade, error) {
	g := &model.Grade{}
	var gradeVal sql.NullFloat64
	var eID, sID sql.NullString
	var eVer, sVer sql.NullInt64
	err := p.db.QueryRowContext(ctx, `
		SELECT g.id, g.grade, g.version,
		       e.id, e.version,
		       p.id, p.version
		FROM grades g
		LEFT JOIN exams e ON e.id = g.exam_id
		LEFT JOIN persons p ON p.id = g.student_id
		WHERE g.id = $1`, id,
	).Scan(&g.ID, &gradeVal, &g.Version, &eID, &eVer, &sID, &sVer)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying grade: %w", err)
	}
	if gradeVal.Valid {
		v := gradeVal.Float64
		g.Grade = &v
	}
	if eID.Valid {
		g.Exam = &model.Exam{ID: eID.String, Version: int(eVer.Int64)}
	}
	if sID.Valid {
		s := &model.Student{}
		s.ID = sID.String
		s.Version = int(sVer.Int64)
		g.Student = s
	}
	return g, nil
}

func (p *Postgres) ListGrades(ctx context.Context) ([]*model.Grade, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT g.id, g.grade, g.version,
		       e.id, e.version,
		       p.id, p.version
		FROM grades g
		LEFT JOIN exams e ON e.id = g.exam_id
		LEFT JOIN persons p ON p.id = g.student_id
		ORDER BY g.id`)
	if err != nil {
		return nil, fmt.Errorf("listing grades: %w", err)
	}
	defer rows.Close()
	var grades []*model.Grade
	for rows.Next() {
		g := &model.Grade{}
		var gradeVal sql.NullFloat64
		var eID, sID sql.NullString
		var eVer, sVer sql.NullInt64
		if err := rows.Scan(&g.ID, &gradeVal, &g.Version, &eID, &eVer, &sID, &sVer); err != nil {
			return nil, err
		}
		if gradeVal.Valid {
			v := gradeVal.Float64
			g.Grade = &v
		}
		if eID.Valid {
			g.Exam = &model.Exam{ID: eID.String, Version: int(eVer.Int64)}
		}
		if sID.Valid {
			s := &model.Student{}
			s.ID = sID.String
			s.Version = int(sVer.Int64)
			g.Student = s
		}
		grades = append(grades, g)
	}
	return grades, rows.Err()
}

func (p *Postgres) UpdateGrade(ctx context.Context, g *model.Grade) error {
	examID, studentID := "", ""
	if g.Exam != nil {
		examID = g.Exam.ID
	}
	if g.Student != nil {
		studentID = g.Student.ID
	}
	res, err := p.db.ExecContext(ctx,
		`UPDATE grades SET grade=$1, exam_id=$2, student_id=$3, version=version+1 WHERE id=$4 AND version=$5`,
		nullableFloat64(g.Grade), nullableString(examID), nullableString(studentID), g.ID, g.Version,
	)
	if err != nil {
		return fmt.Errorf("updating grade: %w", err)
	}
	return optimisticCheck(ctx, p.db, res, "grades", g.ID, &g.Version)
}

func (p *Postgres) DeleteGrade(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "grades", id)
}

func (p *Postgres) IsTeacherOfExam(ctx context.Context, examID, sub string) (bool, error) {
	var exists bool
	err := p.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM exams e
			JOIN school_class_teachers sct ON sct.school_class_id = e.school_class_id
			JOIN persons t ON t.id = sct.teacher_id
			WHERE e.id = $1 AND t.sub = $2
		)`, examID, sub).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("IsTeacherOfExam: %w", err)
	}
	return exists, nil
}

func (p *Postgres) IsTeacherOfGrade(ctx context.Context, gradeID, sub string) (bool, error) {
	var exists bool
	err := p.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM grades g
			JOIN exams e ON e.id = g.exam_id
			JOIN school_class_teachers sct ON sct.school_class_id = e.school_class_id
			JOIN persons t ON t.id = sct.teacher_id
			WHERE g.id = $1 AND t.sub = $2
		)`, gradeID, sub).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("IsTeacherOfGrade: %w", err)
	}
	return exists, nil
}
