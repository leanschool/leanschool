package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	model "github.com/Joel-Haeberli/leanschool-model"
)

// ── shared person helpers ─────────────────────────────────────────────────────

const personJoin = `
	FROM persons p
	LEFT JOIN addresses a ON a.id = p.address_id
	LEFT JOIN postal_codes pc ON pc.id = a.postal_code_id`

const personCols = `p.id, p.name, p.prename, p.date_of_birth, COALESCE(p.sub,''), p.version,
	a.id, a.street, a.number, pc.id, pc.number, pc.city, pc.version`

// scanPersonCore scans the standard person columns into individual fields.
// Column order must match personCols.
func scanPersonCore(row interface {
	Scan(dest ...any) error
}) (id, name, prename string, dob *time.Time, sub *string, version int, addr *model.Address, err error) {
	var dobNull sql.NullTime
	var subStr string
	var addrID, addrStreet, addrNumber sql.NullString
	var pcID sql.NullString
	var pcNo sql.NullInt64
	var pcCity sql.NullString
	var pcVer sql.NullInt64
	if err = row.Scan(
		&id, &name, &prename, &dobNull, &subStr, &version,
		&addrID, &addrStreet, &addrNumber, &pcID, &pcNo, &pcCity, &pcVer,
	); err != nil {
		return
	}
	if dobNull.Valid {
		t := dobNull.Time
		dob = &t
	}
	if subStr != "" {
		sub = &subStr
	}
	if addrID.Valid {
		a := &model.Address{ID: addrID.String}
		if addrStreet.Valid {
			a.Street = &addrStreet.String
		}
		if addrNumber.Valid {
			a.Number = &addrNumber.String
		}
		if pcID.Valid {
			n := int(pcNo.Int64)
			a.PostalCode = &model.PostalCode{
				ID: pcID.String, Number: n, City: pcCity.String, Version: int(pcVer.Int64),
			}
		}
		addr = a
	}
	return
}

// ── Person ────────────────────────────────────────────────────────────────────

func (p *Postgres) CreatePerson(ctx context.Context, person *model.Person) error {
	person.ID = newID()
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := p.createPerson(ctx, tx, person); err != nil {
		return err
	}
	return tx.Commit()
}

func scanPerson(row interface {
	Scan(dest ...any) error
}, person *model.Person) error {
	id, name, prename, dob, sub, version, addr, err := scanPersonCore(row)
	if err != nil {
		return err
	}
	person.ID = id
	person.Name = name
	person.Prename = prename
	person.DateOfBirth = dob
	person.Sub = sub
	person.Version = version
	person.Address = addr
	return nil
}

func (p *Postgres) GetPerson(ctx context.Context, id string) (*model.Person, error) {
	person := &model.Person{}
	err := scanPerson(p.db.QueryRowContext(ctx,
		`SELECT `+personCols+personJoin+` WHERE p.id = $1`, id,
	), person)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying person: %w", err)
	}
	return person, nil
}

func (p *Postgres) ListPersons(ctx context.Context) ([]*model.Person, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT `+personCols+personJoin+` ORDER BY p.name, p.prename`)
	if err != nil {
		return nil, fmt.Errorf("listing persons: %w", err)
	}
	defer rows.Close()
	var persons []*model.Person
	for rows.Next() {
		person := &model.Person{}
		if err := scanPerson(rows, person); err != nil {
			return nil, err
		}
		persons = append(persons, person)
	}
	return persons, rows.Err()
}

func (p *Postgres) UpdatePerson(ctx context.Context, person *model.Person) error {
	addrID := ""
	if person.Address != nil {
		addrID = person.Address.ID
	}
	res, err := p.db.ExecContext(ctx,
		`UPDATE persons SET name=$1, prename=$2, date_of_birth=$3, sub=$4, address_id=$5, version=version+1
		 WHERE id=$6 AND version=$7`,
		person.Name, person.Prename, nullableTime(person.DateOfBirth),
		nullableStringPtr(person.Sub), nullableString(addrID), person.ID, person.Version,
	)
	if err != nil {
		return fmt.Errorf("updating person: %w", err)
	}
	return optimisticCheck(ctx, p.db, res, "persons", person.ID, &person.Version)
}

func (p *Postgres) DeletePerson(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "persons", id)
}

// ── Guardian ──────────────────────────────────────────────────────────────────

func (p *Postgres) CreateGuardian(ctx context.Context, g *model.Guardian) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	g.ID = newID()
	addrID := ""
	if g.Address != nil {
		addrID = g.Address.ID
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO persons (id, name, prename, date_of_birth, sub, address_id, version)
		 VALUES ($1, $2, $3, $4, $5, $6, 0)`,
		g.ID, g.Name, g.Prename, nullableTime(g.DateOfBirth),
		nullableStringPtr(g.Sub), nullableString(addrID),
	); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO guardians (person_id) VALUES ($1)`, g.ID); err != nil {
		return err
	}
	return tx.Commit()
}

func scanGuardian(row interface{ Scan(dest ...any) error }, g *model.Guardian) error {
	id, name, prename, dob, sub, version, addr, err := scanPersonCore(row)
	if err != nil {
		return err
	}
	g.ID = id
	g.Name = name
	g.Prename = prename
	g.DateOfBirth = dob
	g.Sub = sub
	g.Version = version
	g.Address = addr
	return nil
}

func (p *Postgres) GetGuardian(ctx context.Context, id string) (*model.Guardian, error) {
	g := &model.Guardian{}
	err := scanGuardian(p.db.QueryRowContext(ctx,
		`SELECT `+personCols+personJoin+`
		 JOIN guardians gg ON gg.person_id = p.id
		 WHERE p.id = $1`, id,
	), g)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying guardian: %w", err)
	}
	return g, nil
}

func (p *Postgres) ListGuardians(ctx context.Context) ([]*model.Guardian, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT `+personCols+personJoin+`
		 JOIN guardians gg ON gg.person_id = p.id
		 ORDER BY p.name, p.prename`)
	if err != nil {
		return nil, fmt.Errorf("listing guardians: %w", err)
	}
	defer rows.Close()
	var guardians []*model.Guardian
	for rows.Next() {
		g := &model.Guardian{}
		if err := scanGuardian(rows, g); err != nil {
			return nil, err
		}
		guardians = append(guardians, g)
	}
	return guardians, rows.Err()
}

func (p *Postgres) UpdateGuardian(ctx context.Context, g *model.Guardian) error {
	addrID := ""
	if g.Address != nil {
		addrID = g.Address.ID
	}
	res, err := p.db.ExecContext(ctx,
		`UPDATE persons SET name=$1, prename=$2, date_of_birth=$3, sub=$4, address_id=$5, version=version+1
		 WHERE id=$6 AND version=$7`,
		g.Name, g.Prename, nullableTime(g.DateOfBirth),
		nullableStringPtr(g.Sub), nullableString(addrID), g.ID, g.Version,
	)
	if err != nil {
		return fmt.Errorf("updating guardian: %w", err)
	}
	return optimisticCheck(ctx, p.db, res, "persons", g.ID, &g.Version)
}

func (p *Postgres) DeleteGuardian(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "persons", id)
}

// ── Teacher ───────────────────────────────────────────────────────────────────

const teacherCols = personCols + `, t.iban, t.at_school_since, t.at_school_until`
const teacherJoin = personJoin + ` JOIN teachers t ON t.person_id = p.id`

func scanTeacher(row interface{ Scan(dest ...any) error }, t *model.Teacher) error {
	var iban sql.NullString
	var until sql.NullTime
	var dobNull sql.NullTime
	var subStr string
	var addrID, addrStreet, addrNumber sql.NullString
	var pcID sql.NullString
	var pcNo sql.NullInt64
	var pcCity sql.NullString
	var pcVer sql.NullInt64
	if err2 := row.Scan(
		&t.ID, &t.Name, &t.Prename, &dobNull, &subStr, &t.Version,
		&addrID, &addrStreet, &addrNumber, &pcID, &pcNo, &pcCity, &pcVer,
		&iban, &t.AtSchoolSince, &until,
	); err2 != nil {
		return err2
	}
	if dobNull.Valid {
		d := dobNull.Time
		t.DateOfBirth = &d
	}
	if subStr != "" {
		t.Sub = &subStr
	}
	if addrID.Valid {
		a := &model.Address{ID: addrID.String}
		if addrStreet.Valid {
			a.Street = &addrStreet.String
		}
		if addrNumber.Valid {
			a.Number = &addrNumber.String
		}
		if pcID.Valid {
			n := int(pcNo.Int64)
			a.PostalCode = &model.PostalCode{
				ID: pcID.String, Number: n, City: pcCity.String, Version: int(pcVer.Int64),
			}
		}
		t.Address = a
	}
	if iban.Valid && iban.String != "" {
		t.Iban = &iban.String
	}
	if until.Valid {
		u := until.Time
		t.AtSchoolUntil = &u
	}
	return nil
}

func (p *Postgres) CreateTeacher(ctx context.Context, t *model.Teacher) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	t.ID = newID()
	addrID := ""
	if t.Address != nil {
		addrID = t.Address.ID
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO persons (id, name, prename, date_of_birth, sub, address_id, version)
		 VALUES ($1, $2, $3, $4, $5, $6, 0)`,
		t.ID, t.Name, t.Prename, nullableTime(t.DateOfBirth),
		nullableStringPtr(t.Sub), nullableString(addrID),
	); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO teachers (person_id, iban, at_school_since, at_school_until) VALUES ($1, $2, $3, $4)`,
		t.ID, nullableStringPtr(t.Iban), t.AtSchoolSince, nullableTime(t.AtSchoolUntil),
	); err != nil {
		return err
	}
	return tx.Commit()
}

func (p *Postgres) GetTeacher(ctx context.Context, id string) (*model.Teacher, error) {
	t := &model.Teacher{}
	err := scanTeacher(p.db.QueryRowContext(ctx,
		`SELECT `+teacherCols+teacherJoin+` WHERE p.id = $1`, id,
	), t)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying teacher: %w", err)
	}
	return t, nil
}

func (p *Postgres) ListTeachers(ctx context.Context) ([]*model.Teacher, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT `+teacherCols+teacherJoin+` ORDER BY p.name, p.prename`)
	if err != nil {
		return nil, fmt.Errorf("listing teachers: %w", err)
	}
	defer rows.Close()
	var teachers []*model.Teacher
	for rows.Next() {
		t := &model.Teacher{}
		if err := scanTeacher(rows, t); err != nil {
			return nil, err
		}
		teachers = append(teachers, t)
	}
	return teachers, rows.Err()
}

func (p *Postgres) UpdateTeacher(ctx context.Context, t *model.Teacher) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	addrID := ""
	if t.Address != nil {
		addrID = t.Address.ID
	}
	res, err := tx.ExecContext(ctx,
		`UPDATE persons SET name=$1, prename=$2, date_of_birth=$3, sub=$4, address_id=$5, version=version+1
		 WHERE id=$6 AND version=$7`,
		t.Name, t.Prename, nullableTime(t.DateOfBirth),
		nullableStringPtr(t.Sub), nullableString(addrID), t.ID, t.Version,
	)
	if err != nil {
		return fmt.Errorf("updating teacher person: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		var exists bool
		_ = p.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM persons WHERE id=$1)`, t.ID).Scan(&exists)
		if !exists {
			return fmt.Errorf("teacher not found")
		}
		return ErrOptimisticLock
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE teachers SET iban=$1, at_school_since=$2, at_school_until=$3 WHERE person_id=$4`,
		nullableStringPtr(t.Iban), t.AtSchoolSince, nullableTime(t.AtSchoolUntil), t.ID,
	); err != nil {
		return fmt.Errorf("updating teacher fields: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	t.Version++
	return nil
}

func (p *Postgres) DeleteTeacher(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "persons", id)
}

// ── Student ───────────────────────────────────────────────────────────────────

func (p *Postgres) CreateStudent(ctx context.Context, s *model.Student) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	s.ID = newID()
	addrID := ""
	if s.Address != nil {
		addrID = s.Address.ID
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO persons (id, name, prename, date_of_birth, sub, address_id, version)
		 VALUES ($1, $2, $3, $4, $5, $6, 0)`,
		s.ID, s.Name, s.Prename, nullableTime(s.DateOfBirth),
		nullableStringPtr(s.Sub), nullableString(addrID),
	); err != nil {
		return fmt.Errorf("inserting student person: %w", err)
	}

	classID := ""
	if s.CurrentSchoolClass != nil {
		classID = s.CurrentSchoolClass.ID
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO students (person_id, current_school_class_id) VALUES ($1, $2)`,
		s.ID, nullableString(classID),
	); err != nil {
		return fmt.Errorf("inserting student: %w", err)
	}

	for _, g := range s.Guardians {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO student_guardians (student_id, guardian_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			s.ID, g.ID,
		); err != nil {
			return fmt.Errorf("linking guardian: %w", err)
		}
	}
	for _, sc := range s.PastSchoolClasses {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO student_past_classes (student_id, school_class_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			s.ID, sc.ID,
		); err != nil {
			return fmt.Errorf("linking past class: %w", err)
		}
	}
	return tx.Commit()
}

func (p *Postgres) GetStudent(ctx context.Context, id string) (*model.Student, error) {
	s := &model.Student{}
	var classID sql.NullString
	var dobNull sql.NullTime
	var subStr string
	var addrID, addrStreet, addrNumber sql.NullString
	var pcID sql.NullString
	var pcNo sql.NullInt64
	var pcCity sql.NullString
	var pcVer sql.NullInt64
	err := p.db.QueryRowContext(ctx, `
		SELECT p.id, p.name, p.prename, p.date_of_birth, COALESCE(p.sub,''), p.version,
		       a.id, a.street, a.number, pc.id, pc.number, pc.city, pc.version,
		       st.current_school_class_id
		FROM persons p
		LEFT JOIN addresses a ON a.id = p.address_id
		LEFT JOIN postal_codes pc ON pc.id = a.postal_code_id
		JOIN students st ON st.person_id = p.id
		WHERE p.id = $1`, id,
	).Scan(
		&s.ID, &s.Name, &s.Prename, &dobNull, &subStr, &s.Version,
		&addrID, &addrStreet, &addrNumber, &pcID, &pcNo, &pcCity, &pcVer,
		&classID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying student: %w", err)
	}
	if dobNull.Valid {
		d := dobNull.Time
		s.DateOfBirth = &d
	}
	if subStr != "" {
		s.Sub = &subStr
	}
	if addrID.Valid {
		a := &model.Address{ID: addrID.String}
		if addrStreet.Valid {
			a.Street = &addrStreet.String
		}
		if addrNumber.Valid {
			a.Number = &addrNumber.String
		}
		if pcID.Valid {
			n := int(pcNo.Int64)
			a.PostalCode = &model.PostalCode{
				ID: pcID.String, Number: n, City: pcCity.String, Version: int(pcVer.Int64),
			}
		}
		s.Address = a
	}
	if classID.Valid {
		s.CurrentSchoolClass = &model.SchoolClass{ID: classID.String}
	}
	guardians, err := loadStudentGuardians(ctx, p.db, id)
	if err != nil {
		return nil, err
	}
	s.Guardians = guardians

	pastClasses, err := loadStudentPastClasses(ctx, p.db, id)
	if err != nil {
		return nil, err
	}
	s.PastSchoolClasses = pastClasses
	s.Grades = []model.Grade{}
	return s, nil
}

func (p *Postgres) ListStudents(ctx context.Context) ([]*model.Student, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT p.id, p.name, p.prename, p.date_of_birth, COALESCE(p.sub,''), p.version,
		       a.id, a.street, a.number, pc.id, pc.number, pc.city, pc.version,
		       st.current_school_class_id
		FROM persons p
		LEFT JOIN addresses a ON a.id = p.address_id
		LEFT JOIN postal_codes pc ON pc.id = a.postal_code_id
		JOIN students st ON st.person_id = p.id
		ORDER BY p.name, p.prename`)
	if err != nil {
		return nil, fmt.Errorf("listing students: %w", err)
	}
	defer rows.Close()
	var students []*model.Student
	for rows.Next() {
		s := &model.Student{}
		var classID sql.NullString
		var dobNull sql.NullTime
		var subStr string
		var addrID, addrStreet, addrNumber sql.NullString
		var pcID sql.NullString
		var pcNo sql.NullInt64
		var pcCity sql.NullString
		var pcVer sql.NullInt64
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Prename, &dobNull, &subStr, &s.Version,
			&addrID, &addrStreet, &addrNumber, &pcID, &pcNo, &pcCity, &pcVer,
			&classID,
		); err != nil {
			return nil, err
		}
		if dobNull.Valid {
			d := dobNull.Time
			s.DateOfBirth = &d
		}
		if subStr != "" {
			s.Sub = &subStr
		}
		if addrID.Valid {
			a := &model.Address{ID: addrID.String}
			if addrStreet.Valid {
				a.Street = &addrStreet.String
			}
			if addrNumber.Valid {
				a.Number = &addrNumber.String
			}
			if pcID.Valid {
				n := int(pcNo.Int64)
				a.PostalCode = &model.PostalCode{
					ID: pcID.String, Number: n, City: pcCity.String, Version: int(pcVer.Int64),
				}
			}
			s.Address = a
		}
		if classID.Valid {
			s.CurrentSchoolClass = &model.SchoolClass{ID: classID.String}
		}
		s.Guardians = []model.Guardian{}
		s.PastSchoolClasses = []model.SchoolClass{}
		s.Grades = []model.Grade{}
		students = append(students, s)
	}
	return students, rows.Err()
}

func (p *Postgres) UpdateStudent(ctx context.Context, s *model.Student) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	addrID := ""
	if s.Address != nil {
		addrID = s.Address.ID
	}
	res, err := tx.ExecContext(ctx,
		`UPDATE persons SET name=$1, prename=$2, date_of_birth=$3, sub=$4, address_id=$5, version=version+1
		 WHERE id=$6 AND version=$7`,
		s.Name, s.Prename, nullableTime(s.DateOfBirth),
		nullableStringPtr(s.Sub), nullableString(addrID), s.ID, s.Version,
	)
	if err != nil {
		return fmt.Errorf("updating student person: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		var exists bool
		_ = p.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM persons WHERE id=$1)`, s.ID).Scan(&exists)
		if !exists {
			return fmt.Errorf("student not found")
		}
		return ErrOptimisticLock
	}

	classID := ""
	if s.CurrentSchoolClass != nil {
		classID = s.CurrentSchoolClass.ID
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE students SET current_school_class_id=$1 WHERE person_id=$2`,
		nullableString(classID), s.ID,
	); err != nil {
		return fmt.Errorf("updating student fields: %w", err)
	}

	// Replace guardians
	if _, err := tx.ExecContext(ctx, `DELETE FROM student_guardians WHERE student_id=$1`, s.ID); err != nil {
		return fmt.Errorf("clearing student guardians: %w", err)
	}
	for _, g := range s.Guardians {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO student_guardians (student_id, guardian_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			s.ID, g.ID,
		); err != nil {
			return fmt.Errorf("linking guardian: %w", err)
		}
	}

	// Replace past classes
	if _, err := tx.ExecContext(ctx, `DELETE FROM student_past_classes WHERE student_id=$1`, s.ID); err != nil {
		return fmt.Errorf("clearing student past classes: %w", err)
	}
	for _, sc := range s.PastSchoolClasses {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO student_past_classes (student_id, school_class_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			s.ID, sc.ID,
		); err != nil {
			return fmt.Errorf("linking past class: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	s.Version++
	return nil
}

func (p *Postgres) DeleteStudent(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "persons", id)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func loadStudentGuardians(ctx context.Context, db *sql.DB, studentID string) ([]model.Guardian, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT p.id, p.name, p.prename, p.date_of_birth, COALESCE(p.sub,''), p.version,
		       a.id, a.street, a.number, pc.id, pc.number, pc.city, pc.version
		FROM persons p
		LEFT JOIN addresses a ON a.id = p.address_id
		LEFT JOIN postal_codes pc ON pc.id = a.postal_code_id
		JOIN guardians gg ON gg.person_id = p.id
		JOIN student_guardians sg ON sg.guardian_id = p.id
		WHERE sg.student_id = $1
		ORDER BY p.name`, studentID,
	)
	if err != nil {
		return nil, fmt.Errorf("loading student guardians: %w", err)
	}
	defer rows.Close()
	var guardians []model.Guardian
	for rows.Next() {
		g := model.Guardian{}
		if err := scanGuardian(rows, &g); err != nil {
			return nil, err
		}
		guardians = append(guardians, g)
	}
	if guardians == nil {
		guardians = []model.Guardian{}
	}
	return guardians, rows.Err()
}

func loadStudentPastClasses(ctx context.Context, db *sql.DB, studentID string) ([]model.SchoolClass, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT sc.id, sc.name, COALESCE(sc.shortcut,''), sc.version
		FROM school_classes sc
		JOIN student_past_classes spc ON spc.school_class_id = sc.id
		WHERE spc.student_id = $1
		ORDER BY sc.name`, studentID,
	)
	if err != nil {
		return nil, fmt.Errorf("loading student past classes: %w", err)
	}
	defer rows.Close()
	var classes []model.SchoolClass
	for rows.Next() {
		sc := model.SchoolClass{}
		if err := rows.Scan(&sc.ID, &sc.Name, &sc.Shortcut, &sc.Version); err != nil {
			return nil, err
		}
		sc.Teachers = []model.Teacher{}
		sc.Students = []model.Student{}
		classes = append(classes, sc)
	}
	if classes == nil {
		classes = []model.SchoolClass{}
	}
	return classes, rows.Err()
}

func nullableTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t
}
