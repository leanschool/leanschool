package storage

import (
	"context"
	"fmt"

	"github.com/Joel-Haeberli/timetable-service/internal/model"
	"github.com/lib/pq"
)

// ── TeacherSnapshot ──────────────────────────────────────────────────────────

func (p *Postgres) CreateTeacherSnapshot(ctx context.Context, s *model.TeacherSnapshot) error {
	if s.ID == "" {
		s.ID = newID()
	}
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO teacher_snapshots (id, plan_id, name, prename, sub, subjects)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		s.ID, s.PlanID, s.Name, s.Prename, s.Sub, pq.Array(s.Subjects),
	)
	return err
}

func (p *Postgres) ListTeacherSnapshots(ctx context.Context, planID string) ([]*model.TeacherSnapshot, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, plan_id, name, prename, sub, subjects
		FROM teacher_snapshots WHERE plan_id = $1 ORDER BY name, prename`, planID)
	if err != nil {
		return nil, fmt.Errorf("listing teacher snapshots: %w", err)
	}
	defer rows.Close()
	var snapshots []*model.TeacherSnapshot
	for rows.Next() {
		s := &model.TeacherSnapshot{}
		if err := rows.Scan(&s.ID, &s.PlanID, &s.Name, &s.Prename, &s.Sub, pq.Array(&s.Subjects)); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, rows.Err()
}

// ── SubjectSnapshot ──────────────────────────────────────────────────────────

func (p *Postgres) CreateSubjectSnapshot(ctx context.Context, s *model.SubjectSnapshot) error {
	if s.ID == "" {
		s.ID = newID()
	}
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO subject_snapshots (id, plan_id, name) VALUES ($1, $2, $3)`,
		s.ID, s.PlanID, s.Name,
	)
	return err
}

func (p *Postgres) ListSubjectSnapshots(ctx context.Context, planID string) ([]*model.SubjectSnapshot, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, plan_id, name FROM subject_snapshots WHERE plan_id = $1 ORDER BY name`, planID)
	if err != nil {
		return nil, fmt.Errorf("listing subject snapshots: %w", err)
	}
	defer rows.Close()
	var snapshots []*model.SubjectSnapshot
	for rows.Next() {
		s := &model.SubjectSnapshot{}
		if err := rows.Scan(&s.ID, &s.PlanID, &s.Name); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, rows.Err()
}

// ── SchoolClassSnapshot ──────────────────────────────────────────────────────

func (p *Postgres) CreateSchoolClassSnapshot(ctx context.Context, s *model.SchoolClassSnapshot) error {
	if s.ID == "" {
		s.ID = newID()
	}
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO school_class_snapshots (id, plan_id, name, shortcut) VALUES ($1, $2, $3, $4)`,
		s.ID, s.PlanID, s.Name, s.Shortcut,
	)
	return err
}

func (p *Postgres) ListSchoolClassSnapshots(ctx context.Context, planID string) ([]*model.SchoolClassSnapshot, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, plan_id, name, shortcut FROM school_class_snapshots WHERE plan_id = $1 ORDER BY name`, planID)
	if err != nil {
		return nil, fmt.Errorf("listing school class snapshots: %w", err)
	}
	defer rows.Close()
	var snapshots []*model.SchoolClassSnapshot
	for rows.Next() {
		s := &model.SchoolClassSnapshot{}
		if err := rows.Scan(&s.ID, &s.PlanID, &s.Name, &s.Shortcut); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, rows.Err()
}

// ── RoomSnapshot ─────────────────────────────────────────────────────────────

func (p *Postgres) CreateRoomSnapshot(ctx context.Context, s *model.RoomSnapshot) error {
	if s.ID == "" {
		s.ID = newID()
	}
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO room_snapshots (id, plan_id, name, room_type) VALUES ($1, $2, $3, $4)`,
		s.ID, s.PlanID, s.Name, s.RoomType,
	)
	return err
}

func (p *Postgres) ListRoomSnapshots(ctx context.Context, planID string) ([]*model.RoomSnapshot, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, plan_id, name, room_type FROM room_snapshots WHERE plan_id = $1 ORDER BY name`, planID)
	if err != nil {
		return nil, fmt.Errorf("listing room snapshots: %w", err)
	}
	defer rows.Close()
	var snapshots []*model.RoomSnapshot
	for rows.Next() {
		s := &model.RoomSnapshot{}
		if err := rows.Scan(&s.ID, &s.PlanID, &s.Name, &s.RoomType); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, rows.Err()
}

// ── DeleteSnapshotsByPlan ────────────────────────────────────────────────────

func (p *Postgres) DeleteSnapshotsByPlan(ctx context.Context, planID string) error {
	tables := []string{
		"teacher_snapshots",
		"subject_snapshots",
		"school_class_snapshots",
		"room_snapshots",
	}
	for _, table := range tables {
		if _, err := p.db.ExecContext(ctx, fmt.Sprintf(`DELETE FROM %s WHERE plan_id = $1`, table), planID); err != nil {
			return fmt.Errorf("deleting %s: %w", table, err)
		}
	}
	return nil
}
