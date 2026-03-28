package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Joel-Haeberli/timetable-service/internal/model"
)

// ── Conflict CRUD ────────────────────────────────────────────────────────────

func (p *Postgres) CreateConflict(ctx context.Context, c *model.Conflict) error {
	c.ID = newID()

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO conflicts
		(id, plan_id, type, severity, description, teacher_id, school_class_id, time_slot_id, resolved, resolved_by, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 0)`,
		c.ID, c.PlanID, string(c.Type), string(c.Severity), c.Description,
		c.TeacherID, c.SchoolClassID, c.TimeSlotID, c.Resolved, c.ResolvedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting conflict: %w", err)
	}

	for _, entryID := range c.EntryIDs {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO conflict_entries (conflict_id, entry_id) VALUES ($1, $2)`,
			c.ID, entryID,
		)
		if err != nil {
			return fmt.Errorf("inserting conflict entry: %w", err)
		}
	}

	return tx.Commit()
}

func (p *Postgres) GetConflict(ctx context.Context, id string) (*model.Conflict, error) {
	c := &model.Conflict{}
	var cType, severity string
	err := p.db.QueryRowContext(ctx, `
		SELECT id, plan_id, type, severity, description, teacher_id, school_class_id,
		       time_slot_id, resolved, resolved_by, version
		FROM conflicts WHERE id = $1`, id,
	).Scan(&c.ID, &c.PlanID, &cType, &severity, &c.Description,
		&c.TeacherID, &c.SchoolClassID, &c.TimeSlotID, &c.Resolved, &c.ResolvedBy, &c.Version)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying conflict: %w", err)
	}
	c.Type = model.ConflictType(cType)
	c.Severity = model.ConflictSeverity(severity)

	// Load entry IDs from join table
	rows, err := p.db.QueryContext(ctx, `SELECT entry_id FROM conflict_entries WHERE conflict_id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("querying conflict entries: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var entryID string
		if err := rows.Scan(&entryID); err != nil {
			return nil, err
		}
		c.EntryIDs = append(c.EntryIDs, entryID)
	}
	if c.EntryIDs == nil {
		c.EntryIDs = []string{}
	}
	return c, rows.Err()
}

func (p *Postgres) ListConflictsByPlan(ctx context.Context, planID string) ([]*model.Conflict, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, plan_id, type, severity, description, teacher_id, school_class_id,
		       time_slot_id, resolved, resolved_by, version
		FROM conflicts WHERE plan_id = $1 ORDER BY type`, planID)
	if err != nil {
		return nil, fmt.Errorf("listing conflicts: %w", err)
	}
	defer rows.Close()
	var conflicts []*model.Conflict
	for rows.Next() {
		c := &model.Conflict{}
		var cType, severity string
		if err := rows.Scan(&c.ID, &c.PlanID, &cType, &severity, &c.Description,
			&c.TeacherID, &c.SchoolClassID, &c.TimeSlotID, &c.Resolved, &c.ResolvedBy, &c.Version); err != nil {
			return nil, err
		}
		c.Type = model.ConflictType(cType)
		c.Severity = model.ConflictSeverity(severity)
		c.EntryIDs = []string{} // loaded separately if needed
		conflicts = append(conflicts, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Batch load entry IDs for all conflicts
	for _, c := range conflicts {
		entryRows, err := p.db.QueryContext(ctx, `SELECT entry_id FROM conflict_entries WHERE conflict_id = $1`, c.ID)
		if err != nil {
			return nil, fmt.Errorf("querying conflict entries: %w", err)
		}
		for entryRows.Next() {
			var entryID string
			if err := entryRows.Scan(&entryID); err != nil {
				entryRows.Close()
				return nil, err
			}
			c.EntryIDs = append(c.EntryIDs, entryID)
		}
		entryRows.Close()
	}
	return conflicts, nil
}

func (p *Postgres) DeleteConflict(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "conflicts", id)
}

func (p *Postgres) DeleteConflictsByPlan(ctx context.Context, planID string) error {
	_, err := p.db.ExecContext(ctx, `DELETE FROM conflicts WHERE plan_id = $1`, planID)
	return err
}
