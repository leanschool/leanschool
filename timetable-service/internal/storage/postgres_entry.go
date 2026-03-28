package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Joel-Haeberli/timetable-service/internal/model"
)

// ── TimetableEntry CRUD ──────────────────────────────────────────────────────

func (p *Postgres) CreateEntry(ctx context.Context, e *model.TimetableEntry) error {
	e.ID = newID()
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO timetable_entries
		(id, plan_id, school_class_id, school_class_name, subject_id, subject_name,
		 teacher_id, teacher_name, room_id, room_name, time_slot_id, is_double_lesson, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 0)`,
		e.ID, e.PlanID, e.SchoolClassID, e.SchoolClassName, e.SubjectID, e.SubjectName,
		e.TeacherID, e.TeacherName, e.RoomID, e.RoomName, e.TimeSlotID, e.IsDoubleLesson,
	)
	return err
}

func (p *Postgres) BulkCreateEntries(ctx context.Context, entries []*model.TimetableEntry) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO timetable_entries
		(id, plan_id, school_class_id, school_class_name, subject_id, subject_name,
		 teacher_id, teacher_name, room_id, room_name, time_slot_id, is_double_lesson, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 0)`)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, e := range entries {
		e.ID = newID()
		if _, err := stmt.ExecContext(ctx,
			e.ID, e.PlanID, e.SchoolClassID, e.SchoolClassName, e.SubjectID, e.SubjectName,
			e.TeacherID, e.TeacherName, e.RoomID, e.RoomName, e.TimeSlotID, e.IsDoubleLesson,
		); err != nil {
			return fmt.Errorf("inserting entry: %w", err)
		}
	}
	return tx.Commit()
}

func (p *Postgres) GetEntry(ctx context.Context, id string) (*model.TimetableEntry, error) {
	e := &model.TimetableEntry{}
	err := p.db.QueryRowContext(ctx, `
		SELECT id, plan_id, school_class_id, school_class_name, subject_id, subject_name,
		       teacher_id, teacher_name, room_id, room_name, time_slot_id, is_double_lesson, version
		FROM timetable_entries WHERE id = $1`, id,
	).Scan(&e.ID, &e.PlanID, &e.SchoolClassID, &e.SchoolClassName, &e.SubjectID, &e.SubjectName,
		&e.TeacherID, &e.TeacherName, &e.RoomID, &e.RoomName, &e.TimeSlotID, &e.IsDoubleLesson, &e.Version)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying entry: %w", err)
	}
	return e, nil
}

func (p *Postgres) ListEntriesByPlan(ctx context.Context, planID string) ([]*model.TimetableEntry, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, plan_id, school_class_id, school_class_name, subject_id, subject_name,
		       teacher_id, teacher_name, room_id, room_name, time_slot_id, is_double_lesson, version
		FROM timetable_entries WHERE plan_id = $1
		ORDER BY school_class_id, time_slot_id`, planID)
	if err != nil {
		return nil, fmt.Errorf("listing entries: %w", err)
	}
	defer rows.Close()
	var entries []*model.TimetableEntry
	for rows.Next() {
		e := &model.TimetableEntry{}
		if err := rows.Scan(&e.ID, &e.PlanID, &e.SchoolClassID, &e.SchoolClassName, &e.SubjectID, &e.SubjectName,
			&e.TeacherID, &e.TeacherName, &e.RoomID, &e.RoomName, &e.TimeSlotID, &e.IsDoubleLesson, &e.Version); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (p *Postgres) UpdateEntry(ctx context.Context, e *model.TimetableEntry) error {
	res, err := p.db.ExecContext(ctx, `
		UPDATE timetable_entries
		SET school_class_id=$1, school_class_name=$2, subject_id=$3, subject_name=$4,
		    teacher_id=$5, teacher_name=$6, room_id=$7, room_name=$8,
		    time_slot_id=$9, is_double_lesson=$10, version=version+1
		WHERE id=$11 AND version=$12`,
		e.SchoolClassID, e.SchoolClassName, e.SubjectID, e.SubjectName,
		e.TeacherID, e.TeacherName, e.RoomID, e.RoomName,
		e.TimeSlotID, e.IsDoubleLesson,
		e.ID, e.Version,
	)
	if err != nil {
		return fmt.Errorf("updating entry: %w", err)
	}
	return optimisticCheck(ctx, p.db, res, "timetable_entries", e.ID, &e.Version)
}

func (p *Postgres) DeleteEntry(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "timetable_entries", id)
}

func (p *Postgres) DeleteEntriesByPlan(ctx context.Context, planID string) error {
	_, err := p.db.ExecContext(ctx, `DELETE FROM timetable_entries WHERE plan_id = $1`, planID)
	return err
}
