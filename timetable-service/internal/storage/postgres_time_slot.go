package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Joel-Haeberli/timetable-service/internal/model"
)

// ── TimeSlotDefinition CRUD ──────────────────────────────────────────────────

func (p *Postgres) CreateTimeSlot(ctx context.Context, ts *model.TimeSlotDefinition) error {
	ts.ID = newID()
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO time_slot_definitions (id, plan_id, day_of_week, period, start_time, end_time, is_morning, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 0)`,
		ts.ID, ts.PlanID, int(ts.DayOfWeek), ts.Period, ts.StartTime, ts.EndTime, ts.IsMorning,
	)
	return err
}

func (p *Postgres) GetTimeSlot(ctx context.Context, id string) (*model.TimeSlotDefinition, error) {
	ts := &model.TimeSlotDefinition{}
	var dow int
	err := p.db.QueryRowContext(ctx, `
		SELECT id, plan_id, day_of_week, period, start_time, end_time, is_morning, version
		FROM time_slot_definitions WHERE id = $1`, id,
	).Scan(&ts.ID, &ts.PlanID, &dow, &ts.Period, &ts.StartTime, &ts.EndTime, &ts.IsMorning, &ts.Version)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying time slot: %w", err)
	}
	ts.DayOfWeek = model.DayOfWeek(dow)
	return ts, nil
}

func (p *Postgres) ListTimeSlotsByPlan(ctx context.Context, planID string) ([]*model.TimeSlotDefinition, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, plan_id, day_of_week, period, start_time, end_time, is_morning, version
		FROM time_slot_definitions WHERE plan_id = $1
		ORDER BY day_of_week, period`, planID)
	if err != nil {
		return nil, fmt.Errorf("listing time slots: %w", err)
	}
	defer rows.Close()
	var slots []*model.TimeSlotDefinition
	for rows.Next() {
		ts := &model.TimeSlotDefinition{}
		var dow int
		if err := rows.Scan(&ts.ID, &ts.PlanID, &dow, &ts.Period, &ts.StartTime, &ts.EndTime, &ts.IsMorning, &ts.Version); err != nil {
			return nil, err
		}
		ts.DayOfWeek = model.DayOfWeek(dow)
		slots = append(slots, ts)
	}
	return slots, rows.Err()
}

func (p *Postgres) UpdateTimeSlot(ctx context.Context, ts *model.TimeSlotDefinition) error {
	res, err := p.db.ExecContext(ctx, `
		UPDATE time_slot_definitions
		SET day_of_week=$1, period=$2, start_time=$3, end_time=$4, is_morning=$5, version=version+1
		WHERE id=$6 AND version=$7`,
		int(ts.DayOfWeek), ts.Period, ts.StartTime, ts.EndTime, ts.IsMorning, ts.ID, ts.Version,
	)
	if err != nil {
		return fmt.Errorf("updating time slot: %w", err)
	}
	return optimisticCheck(ctx, p.db, res, "time_slot_definitions", ts.ID, &ts.Version)
}

func (p *Postgres) DeleteTimeSlot(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "time_slot_definitions", id)
}

func (p *Postgres) DeleteTimeSlotsByPlan(ctx context.Context, planID string) error {
	_, err := p.db.ExecContext(ctx, `DELETE FROM time_slot_definitions WHERE plan_id = $1`, planID)
	return err
}
