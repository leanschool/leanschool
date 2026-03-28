package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Postgres implements Storage using PostgreSQL.
type Postgres struct {
	db *sql.DB
}

// NewPostgres opens a connection to the given DSN and runs schema migrations.
func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("pinging db: %w", err)
	}
	p := &Postgres{db: db}
	if err := p.migrate(ctx); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}
	return p, nil
}

func (p *Postgres) migrate(ctx context.Context) error {
	_, err := p.db.ExecContext(ctx, `
		-- ── timetable plans ──────────────────────────────────────────────────────
		CREATE TABLE IF NOT EXISTS timetable_plans (
			id             TEXT PRIMARY KEY,
			school_year_id TEXT NOT NULL,
			name           TEXT NOT NULL,
			status         TEXT NOT NULL DEFAULT 'draft',
			created_by     TEXT NOT NULL DEFAULT '',
			created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			version        INT NOT NULL DEFAULT 0
		);

		-- ── time slot definitions ────────────────────────────────────────────────
		CREATE TABLE IF NOT EXISTS time_slot_definitions (
			id          TEXT PRIMARY KEY,
			plan_id     TEXT NOT NULL REFERENCES timetable_plans(id) ON DELETE CASCADE,
			day_of_week INT  NOT NULL,
			period      INT  NOT NULL,
			start_time  TEXT NOT NULL,
			end_time    TEXT NOT NULL,
			is_morning  BOOLEAN NOT NULL DEFAULT TRUE,
			version     INT NOT NULL DEFAULT 0,
			UNIQUE (plan_id, day_of_week, period)
		);
		CREATE INDEX IF NOT EXISTS idx_time_slots_plan ON time_slot_definitions(plan_id);

		-- ── grade requirements ───────────────────────────────────────────────────
		CREATE TABLE IF NOT EXISTS grade_requirements (
			id                  TEXT PRIMARY KEY,
			plan_id             TEXT NOT NULL REFERENCES timetable_plans(id) ON DELETE CASCADE,
			school_class_id     TEXT NOT NULL,
			subject_id          TEXT NOT NULL,
			subject_name        TEXT NOT NULL DEFAULT '',
			lessons_per_week    INT  NOT NULL DEFAULT 0,
			max_double_lessons  INT  NOT NULL DEFAULT 0,
			prefer_morning      BOOLEAN NOT NULL DEFAULT FALSE,
			lesson_duration_min INT  NOT NULL DEFAULT 45,
			version             INT  NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_requirements_plan ON grade_requirements(plan_id);

		-- ── class constraints ────────────────────────────────────────────────────
		CREATE TABLE IF NOT EXISTS class_constraints (
			id                  TEXT PRIMARY KEY,
			plan_id             TEXT NOT NULL REFERENCES timetable_plans(id) ON DELETE CASCADE,
			school_class_id     TEXT NOT NULL,
			school_class_name   TEXT NOT NULL DEFAULT '',
			max_early_starts    INT  NOT NULL DEFAULT 0,
			morning_periods     INT  NOT NULL DEFAULT 4,
			afternoon_periods   INT  NOT NULL DEFAULT 3,
			free_afternoons     INT  NOT NULL DEFAULT 0,
			free_afternoon_days TEXT NOT NULL DEFAULT '',
			has_timetable       BOOLEAN NOT NULL DEFAULT TRUE,
			version             INT  NOT NULL DEFAULT 0,
			UNIQUE (plan_id, school_class_id)
		);
		CREATE INDEX IF NOT EXISTS idx_constraints_plan ON class_constraints(plan_id);

		-- ── timetable entries ────────────────────────────────────────────────────
		CREATE TABLE IF NOT EXISTS timetable_entries (
			id               TEXT PRIMARY KEY,
			plan_id          TEXT NOT NULL REFERENCES timetable_plans(id) ON DELETE CASCADE,
			school_class_id  TEXT NOT NULL,
			school_class_name TEXT NOT NULL DEFAULT '',
			subject_id       TEXT NOT NULL,
			subject_name     TEXT NOT NULL DEFAULT '',
			teacher_id       TEXT NOT NULL DEFAULT '',
			teacher_name     TEXT NOT NULL DEFAULT '',
			room_id          TEXT NOT NULL DEFAULT '',
			room_name        TEXT NOT NULL DEFAULT '',
			time_slot_id     TEXT NOT NULL,
			is_double_lesson BOOLEAN NOT NULL DEFAULT FALSE,
			version          INT  NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_entries_plan ON timetable_entries(plan_id);
		CREATE INDEX IF NOT EXISTS idx_entries_teacher ON timetable_entries(teacher_id);
		CREATE INDEX IF NOT EXISTS idx_entries_class ON timetable_entries(school_class_id);

		-- ── conflicts ────────────────────────────────────────────────────────────
		CREATE TABLE IF NOT EXISTS conflicts (
			id             TEXT PRIMARY KEY,
			plan_id        TEXT NOT NULL REFERENCES timetable_plans(id) ON DELETE CASCADE,
			type           TEXT NOT NULL,
			severity       TEXT NOT NULL DEFAULT 'error',
			description    TEXT NOT NULL DEFAULT '',
			teacher_id     TEXT NOT NULL DEFAULT '',
			school_class_id TEXT NOT NULL DEFAULT '',
			time_slot_id   TEXT NOT NULL DEFAULT '',
			resolved       BOOLEAN NOT NULL DEFAULT FALSE,
			resolved_by    TEXT NOT NULL DEFAULT '',
			version        INT  NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_conflicts_plan ON conflicts(plan_id);

		-- ── conflict ↔ entry join table ──────────────────────────────────────────
		CREATE TABLE IF NOT EXISTS conflict_entries (
			conflict_id TEXT NOT NULL REFERENCES conflicts(id) ON DELETE CASCADE,
			entry_id    TEXT NOT NULL REFERENCES timetable_entries(id) ON DELETE CASCADE,
			PRIMARY KEY (conflict_id, entry_id)
		);

		-- ── snapshots ────────────────────────────────────────────────────────────
		CREATE TABLE IF NOT EXISTS teacher_snapshots (
			id       TEXT PRIMARY KEY,
			plan_id  TEXT NOT NULL REFERENCES timetable_plans(id) ON DELETE CASCADE,
			name     TEXT NOT NULL,
			prename  TEXT NOT NULL DEFAULT '',
			sub      TEXT NOT NULL DEFAULT '',
			subjects TEXT[] NOT NULL DEFAULT '{}'
		);
		CREATE INDEX IF NOT EXISTS idx_teacher_snap_plan ON teacher_snapshots(plan_id);

		CREATE TABLE IF NOT EXISTS subject_snapshots (
			id      TEXT PRIMARY KEY,
			plan_id TEXT NOT NULL REFERENCES timetable_plans(id) ON DELETE CASCADE,
			name    TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_subject_snap_plan ON subject_snapshots(plan_id);

		CREATE TABLE IF NOT EXISTS school_class_snapshots (
			id       TEXT PRIMARY KEY,
			plan_id  TEXT NOT NULL REFERENCES timetable_plans(id) ON DELETE CASCADE,
			name     TEXT NOT NULL,
			shortcut TEXT NOT NULL DEFAULT ''
		);
		CREATE INDEX IF NOT EXISTS idx_class_snap_plan ON school_class_snapshots(plan_id);

		CREATE TABLE IF NOT EXISTS room_snapshots (
			id      TEXT PRIMARY KEY,
			plan_id TEXT NOT NULL REFERENCES timetable_plans(id) ON DELETE CASCADE,
			name    TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_room_snap_plan ON room_snapshots(plan_id);

		ALTER TABLE room_snapshots ADD COLUMN IF NOT EXISTS room_type TEXT NOT NULL DEFAULT '';
	`)
	return err
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func optimisticCheck(ctx context.Context, db *sql.DB, res sql.Result, table, id string, version *int) error {
	n, _ := res.RowsAffected()
	if n == 0 {
		var exists bool
		_ = db.QueryRowContext(ctx,
			fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE id=$1)`, table), id,
		).Scan(&exists)
		if !exists {
			return fmt.Errorf("%s not found", table)
		}
		return ErrOptimisticLock
	}
	*version++
	return nil
}

func deleteByID(ctx context.Context, db *sql.DB, table, id string) error {
	res, err := db.ExecContext(ctx,
		fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, table), id,
	)
	if err != nil {
		return fmt.Errorf("deleting from %s: %w", table, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("%s not found", table)
	}
	return nil
}
