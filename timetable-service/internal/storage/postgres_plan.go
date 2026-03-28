package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Joel-Haeberli/timetable-service/internal/model"
)

// ── TimetablePlan CRUD ───────────────────────────────────────────────────────

func (p *Postgres) CreatePlan(ctx context.Context, plan *model.TimetablePlan) error {
	plan.ID = newID()
	now := time.Now()
	plan.CreatedAt = &now
	plan.UpdatedAt = &now
	if plan.Status == "" {
		plan.Status = model.PlanStatusDraft
	}
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO timetable_plans (id, school_year_id, name, status, created_by, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 0)`,
		plan.ID, plan.SchoolYearID, plan.Name, string(plan.Status), plan.CreatedBy, plan.CreatedAt, plan.UpdatedAt,
	)
	return err
}

func (p *Postgres) GetPlan(ctx context.Context, id string) (*model.TimetablePlan, error) {
	plan := &model.TimetablePlan{}
	var status string
	err := p.db.QueryRowContext(ctx, `
		SELECT id, school_year_id, name, status, created_by, created_at, updated_at, version
		FROM timetable_plans WHERE id = $1`, id,
	).Scan(&plan.ID, &plan.SchoolYearID, &plan.Name, &status, &plan.CreatedBy, &plan.CreatedAt, &plan.UpdatedAt, &plan.Version)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying plan: %w", err)
	}
	plan.Status = model.PlanStatus(status)
	return plan, nil
}

func (p *Postgres) ListPlans(ctx context.Context) ([]*model.TimetablePlan, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, school_year_id, name, status, created_by, created_at, updated_at, version
		FROM timetable_plans ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("listing plans: %w", err)
	}
	defer rows.Close()
	var plans []*model.TimetablePlan
	for rows.Next() {
		plan := &model.TimetablePlan{}
		var status string
		if err := rows.Scan(&plan.ID, &plan.SchoolYearID, &plan.Name, &status, &plan.CreatedBy, &plan.CreatedAt, &plan.UpdatedAt, &plan.Version); err != nil {
			return nil, err
		}
		plan.Status = model.PlanStatus(status)
		plans = append(plans, plan)
	}
	return plans, rows.Err()
}

func (p *Postgres) UpdatePlan(ctx context.Context, plan *model.TimetablePlan) error {
	now := time.Now()
	plan.UpdatedAt = &now
	res, err := p.db.ExecContext(ctx, `
		UPDATE timetable_plans
		SET school_year_id=$1, name=$2, status=$3, updated_at=$4, version=version+1
		WHERE id=$5 AND version=$6`,
		plan.SchoolYearID, plan.Name, string(plan.Status), plan.UpdatedAt, plan.ID, plan.Version,
	)
	if err != nil {
		return fmt.Errorf("updating plan: %w", err)
	}
	return optimisticCheck(ctx, p.db, res, "timetable_plans", plan.ID, &plan.Version)
}

func (p *Postgres) DeletePlan(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "timetable_plans", id)
}
