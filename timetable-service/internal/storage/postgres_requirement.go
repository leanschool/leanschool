package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Joel-Haeberli/timetable-service/internal/model"
)

// ── GradeRequirement CRUD ────────────────────────────────────────────────────

func (p *Postgres) CreateRequirement(ctx context.Context, r *model.GradeRequirement) error {
	r.ID = newID()
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO grade_requirements
		(id, plan_id, school_class_id, subject_id, subject_name, lessons_per_week, max_double_lessons, prefer_morning, lesson_duration_min, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 0)`,
		r.ID, r.PlanID, r.SchoolClassID, r.SubjectID, r.SubjectName,
		r.LessonsPerWeek, r.MaxDoubleLessons, r.PreferMorning, r.LessonDurationMin,
	)
	return err
}

func (p *Postgres) GetRequirement(ctx context.Context, id string) (*model.GradeRequirement, error) {
	r := &model.GradeRequirement{}
	err := p.db.QueryRowContext(ctx, `
		SELECT id, plan_id, school_class_id, subject_id, subject_name,
		       lessons_per_week, max_double_lessons, prefer_morning, lesson_duration_min, version
		FROM grade_requirements WHERE id = $1`, id,
	).Scan(&r.ID, &r.PlanID, &r.SchoolClassID, &r.SubjectID, &r.SubjectName,
		&r.LessonsPerWeek, &r.MaxDoubleLessons, &r.PreferMorning, &r.LessonDurationMin, &r.Version)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying requirement: %w", err)
	}
	return r, nil
}

func (p *Postgres) ListRequirementsByPlan(ctx context.Context, planID string) ([]*model.GradeRequirement, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, plan_id, school_class_id, subject_id, subject_name,
		       lessons_per_week, max_double_lessons, prefer_morning, lesson_duration_min, version
		FROM grade_requirements WHERE plan_id = $1 ORDER BY school_class_id, subject_id`, planID)
	if err != nil {
		return nil, fmt.Errorf("listing requirements: %w", err)
	}
	defer rows.Close()
	var reqs []*model.GradeRequirement
	for rows.Next() {
		r := &model.GradeRequirement{}
		if err := rows.Scan(&r.ID, &r.PlanID, &r.SchoolClassID, &r.SubjectID, &r.SubjectName,
			&r.LessonsPerWeek, &r.MaxDoubleLessons, &r.PreferMorning, &r.LessonDurationMin, &r.Version); err != nil {
			return nil, err
		}
		reqs = append(reqs, r)
	}
	return reqs, rows.Err()
}

func (p *Postgres) UpdateRequirement(ctx context.Context, r *model.GradeRequirement) error {
	res, err := p.db.ExecContext(ctx, `
		UPDATE grade_requirements
		SET school_class_id=$1, subject_id=$2, subject_name=$3,
		    lessons_per_week=$4, max_double_lessons=$5, prefer_morning=$6, lesson_duration_min=$7,
		    version=version+1
		WHERE id=$8 AND version=$9`,
		r.SchoolClassID, r.SubjectID, r.SubjectName,
		r.LessonsPerWeek, r.MaxDoubleLessons, r.PreferMorning, r.LessonDurationMin,
		r.ID, r.Version,
	)
	if err != nil {
		return fmt.Errorf("updating requirement: %w", err)
	}
	return optimisticCheck(ctx, p.db, res, "grade_requirements", r.ID, &r.Version)
}

func (p *Postgres) DeleteRequirement(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "grade_requirements", id)
}
