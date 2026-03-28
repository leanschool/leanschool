package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Joel-Haeberli/timetable-service/internal/model"
)

// ── ClassConstraint CRUD ─────────────────────────────────────────────────────

func (p *Postgres) CreateConstraint(ctx context.Context, c *model.ClassConstraint) error {
	c.ID = newID()
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO class_constraints
		(id, plan_id, school_class_id, school_class_name, max_early_starts,
		 morning_periods, afternoon_periods, free_afternoons, free_afternoon_days, has_timetable, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 0)`,
		c.ID, c.PlanID, c.SchoolClassID, c.SchoolClassName, c.MaxEarlyStarts,
		c.MorningPeriods, c.AfternoonPeriods, c.FreeAfternoons, c.FreeAfternoonDays, c.HasTimetable,
	)
	return err
}

func (p *Postgres) GetConstraint(ctx context.Context, id string) (*model.ClassConstraint, error) {
	c := &model.ClassConstraint{}
	err := p.db.QueryRowContext(ctx, `
		SELECT id, plan_id, school_class_id, school_class_name, max_early_starts,
		       morning_periods, afternoon_periods, free_afternoons, free_afternoon_days, has_timetable, version
		FROM class_constraints WHERE id = $1`, id,
	).Scan(&c.ID, &c.PlanID, &c.SchoolClassID, &c.SchoolClassName, &c.MaxEarlyStarts,
		&c.MorningPeriods, &c.AfternoonPeriods, &c.FreeAfternoons, &c.FreeAfternoonDays, &c.HasTimetable, &c.Version)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying constraint: %w", err)
	}
	return c, nil
}

func (p *Postgres) ListConstraintsByPlan(ctx context.Context, planID string) ([]*model.ClassConstraint, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, plan_id, school_class_id, school_class_name, max_early_starts,
		       morning_periods, afternoon_periods, free_afternoons, free_afternoon_days, has_timetable, version
		FROM class_constraints WHERE plan_id = $1 ORDER BY school_class_name`, planID)
	if err != nil {
		return nil, fmt.Errorf("listing constraints: %w", err)
	}
	defer rows.Close()
	var constraints []*model.ClassConstraint
	for rows.Next() {
		c := &model.ClassConstraint{}
		if err := rows.Scan(&c.ID, &c.PlanID, &c.SchoolClassID, &c.SchoolClassName, &c.MaxEarlyStarts,
			&c.MorningPeriods, &c.AfternoonPeriods, &c.FreeAfternoons, &c.FreeAfternoonDays, &c.HasTimetable, &c.Version); err != nil {
			return nil, err
		}
		constraints = append(constraints, c)
	}
	return constraints, rows.Err()
}

func (p *Postgres) UpdateConstraint(ctx context.Context, c *model.ClassConstraint) error {
	res, err := p.db.ExecContext(ctx, `
		UPDATE class_constraints
		SET school_class_id=$1, school_class_name=$2, max_early_starts=$3,
		    morning_periods=$4, afternoon_periods=$5, free_afternoons=$6,
		    free_afternoon_days=$7, has_timetable=$8, version=version+1
		WHERE id=$9 AND version=$10`,
		c.SchoolClassID, c.SchoolClassName, c.MaxEarlyStarts,
		c.MorningPeriods, c.AfternoonPeriods, c.FreeAfternoons,
		c.FreeAfternoonDays, c.HasTimetable,
		c.ID, c.Version,
	)
	if err != nil {
		return fmt.Errorf("updating constraint: %w", err)
	}
	return optimisticCheck(ctx, p.db, res, "class_constraints", c.ID, &c.Version)
}

func (p *Postgres) DeleteConstraint(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "class_constraints", id)
}
