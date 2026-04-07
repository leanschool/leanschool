package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/lib/pq"
)

// Helper function to get address ID safely
func getAddressID(addr *model.Address) string {
	if addr != nil {
		return addr.ID
	}
	return ""
}

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
		-- ── Step 1: tables with no cross-dependencies ─────────────────────────────

		CREATE TABLE IF NOT EXISTS locations (
			id      TEXT PRIMARY KEY,
			lon     NUMERIC(12,8),
			lat     NUMERIC(12,8),
			version INT NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS buildings (
			id          TEXT PRIMARY KEY,
			name        TEXT NOT NULL,
			location_id TEXT REFERENCES locations(id),
			version     INT NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS rooms (
			id          TEXT PRIMARY KEY,
			name        TEXT NOT NULL,
			building_id TEXT REFERENCES buildings(id),
			version     INT NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS postal_codes (
			number  INT  PRIMARY KEY,
			city    TEXT NOT NULL,
			version INT  NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS cities (
			id      TEXT PRIMARY KEY,
			name    TEXT NOT NULL,
			version INT  NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS city_postal_codes (
			city_id        TEXT NOT NULL REFERENCES cities(id) ON DELETE CASCADE,
			postal_code_no INT  NOT NULL REFERENCES postal_codes(number) ON DELETE CASCADE,
			PRIMARY KEY (city_id, postal_code_no)
		);

		CREATE TABLE IF NOT EXISTS addresses (
			id             TEXT PRIMARY KEY,
			street         TEXT,
			number         TEXT,
			postal_code_no INT REFERENCES postal_codes(number),
			version        INT NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS school_years (
			id      TEXT PRIMARY KEY,
			from_dt TIMESTAMPTZ,
			to_dt   TIMESTAMPTZ,
			version INT NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS curricula (
			id           TEXT PRIMARY KEY,
			name         TEXT,
			active_since TIMESTAMPTZ,
			active_until TIMESTAMPTZ,
			active_from  TIMESTAMPTZ,
			version      INT NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS subjects (
			id      TEXT PRIMARY KEY,
			name    TEXT,
			version INT NOT NULL DEFAULT 0
		);

		-- ── Step 2: legacy domain tables ──────────────────────────────────────────

		CREATE TABLE IF NOT EXISTS accounts (
			id         TEXT          PRIMARY KEY,
			name       TEXT          NOT NULL,
			shortcut   TEXT          NOT NULL,
			balance    NUMERIC(12,2) NOT NULL DEFAULT 0
		);

		DO $$ BEGIN
			ALTER TABLE accounts RENAME COLUMN balance TO budget;
		EXCEPTION WHEN undefined_column THEN NULL;
		END $$;

		ALTER TABLE accounts ADD COLUMN IF NOT EXISTS valid_from DATE NOT NULL DEFAULT (DATE_TRUNC('year', CURRENT_DATE)::DATE);
		ALTER TABLE accounts ADD COLUMN IF NOT EXISTS valid_to   DATE NOT NULL DEFAULT ((DATE_TRUNC('year', CURRENT_DATE) + INTERVAL '1 year' - INTERVAL '1 day')::DATE);

		CREATE TABLE IF NOT EXISTS persons (
			id   TEXT PRIMARY KEY,
			name TEXT NOT NULL
		);
		ALTER TABLE persons ADD COLUMN IF NOT EXISTS sub TEXT;

		CREATE TABLE IF NOT EXISTS receipts (
			id               TEXT          PRIMARY KEY,
			receipt_owner_id TEXT          NOT NULL,
			total_price      NUMERIC(12,2) NOT NULL,
			taxes            NUMERIC(12,2) NOT NULL,
			time             TIMESTAMPTZ   NOT NULL,
			account_id       TEXT          REFERENCES accounts(id),
			status           TEXT          NOT NULL DEFAULT 'unsubmitted'
		);

		CREATE TABLE IF NOT EXISTS receipt_items (
			id         SERIAL PRIMARY KEY,
			receipt_id TEXT          NOT NULL REFERENCES receipts(id) ON DELETE CASCADE,
			name       TEXT          NOT NULL,
			amount     NUMERIC(10,3) NOT NULL,
			price      NUMERIC(12,2) NOT NULL
		);

		ALTER TABLE receipts ADD COLUMN IF NOT EXISTS account_id TEXT REFERENCES accounts(id);
		ALTER TABLE receipts ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'unsubmitted';
		ALTER TABLE receipts ADD COLUMN IF NOT EXISTS file_id TEXT;

		CREATE TABLE IF NOT EXISTS registration_requests (
			id            TEXT        PRIMARY KEY,
			user_sub      TEXT        NOT NULL UNIQUE,
			email         TEXT        NOT NULL,
			desired_roles TEXT[]      NOT NULL,
			status        TEXT        NOT NULL DEFAULT 'pending',
			created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS user_profiles (
			user_sub         TEXT    PRIMARY KEY,
			iban             TEXT    NOT NULL DEFAULT '',
			address          TEXT    NOT NULL DEFAULT '',
			phone            TEXT    NOT NULL DEFAULT '',
			profile_complete BOOLEAN NOT NULL DEFAULT FALSE,
			profile_skipped  BOOLEAN NOT NULL DEFAULT FALSE,
			updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		-- New User Management Tables
		CREATE TABLE IF NOT EXISTS user_registry (
			id                TEXT PRIMARY KEY,
			user_sub          TEXT UNIQUE,
			person_id         TEXT REFERENCES persons(id),
			email             TEXT,
			registration_status TEXT NOT NULL DEFAULT 'pending',
			keycloak_roles    TEXT[] NOT NULL DEFAULT '{}',
			local_roles       TEXT[] NOT NULL DEFAULT '{}',
			last_role_sync    TIMESTAMPTZ,
			created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			last_login_at     TIMESTAMPTZ,
			created_by        TEXT REFERENCES user_registry(id),
			updated_by        TEXT REFERENCES user_registry(id),
			legacy_profile_id TEXT REFERENCES user_profiles(user_sub),
			legacy_request_id TEXT REFERENCES registration_requests(user_sub)
		);

		CREATE TABLE IF NOT EXISTS teachers (
			person_id       TEXT PRIMARY KEY REFERENCES persons(id) ON DELETE CASCADE,
			iban            TEXT,
			at_school_since TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			at_school_until TIMESTAMPTZ
		);

		CREATE TABLE IF NOT EXISTS school_classes (
			id   TEXT PRIMARY KEY,
			name TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS students (
			person_id               TEXT PRIMARY KEY REFERENCES persons(id) ON DELETE CASCADE,
			current_school_class_id TEXT REFERENCES school_classes(id)
		);

		CREATE TABLE IF NOT EXISTS guardians (
			person_id TEXT PRIMARY KEY REFERENCES persons(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS student_guardians (
			student_id  TEXT NOT NULL REFERENCES students(person_id) ON DELETE CASCADE,
			guardian_id TEXT NOT NULL REFERENCES guardians(person_id) ON DELETE CASCADE,
			PRIMARY KEY (student_id, guardian_id)
		);

		CREATE TABLE IF NOT EXISTS registration_workflow (
			id                   TEXT PRIMARY KEY,
			user_id              TEXT REFERENCES user_registry(id),
			request_type         TEXT NOT NULL,
			request_data         JSONB NOT NULL,
			desired_roles        TEXT[] NOT NULL,
			current_roles        TEXT[] NOT NULL DEFAULT '{}',
			approval_status      TEXT NOT NULL DEFAULT 'pending',
			approval_by          TEXT,
			approval_at          TIMESTAMPTZ,
			approval_notes       TEXT,
			rejection_reason     TEXT,
			rejection_by         TEXT,
			rejection_at         TIMESTAMPTZ,
			requires_manual_approval BOOLEAN NOT NULL DEFAULT TRUE,
			auto_approval_rules  JSONB,
			pending_teacher_id   TEXT REFERENCES teachers(person_id),
			pending_student_id   TEXT REFERENCES students(person_id),
			pending_guardian_id  TEXT REFERENCES guardians(person_id),
			created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			completed_at         TIMESTAMPTZ
		);

		CREATE TABLE IF NOT EXISTS role_mappings (
			keycloak_role     TEXT PRIMARY KEY,
			local_role        TEXT NOT NULL UNIQUE,
			description       TEXT NOT NULL,
			auto_assign       BOOLEAN NOT NULL DEFAULT FALSE,
			auto_create       BOOLEAN NOT NULL DEFAULT FALSE,
			requires_approval BOOLEAN NOT NULL DEFAULT TRUE,
			can_teach         BOOLEAN NOT NULL DEFAULT FALSE,
			can_manage_classes BOOLEAN NOT NULL DEFAULT FALSE,
			can_manage_users  BOOLEAN NOT NULL DEFAULT FALSE,
			can_view_reports  BOOLEAN NOT NULL DEFAULT FALSE,
			show_in_registration BOOLEAN NOT NULL DEFAULT TRUE,
			registration_order INT NOT NULL DEFAULT 0,
			created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS school_classes (
			id   TEXT PRIMARY KEY,
			name TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS class_persons (
			class_id  TEXT NOT NULL REFERENCES school_classes(id) ON DELETE CASCADE,
			person_id TEXT NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
			role      TEXT NOT NULL,
			PRIMARY KEY (class_id, person_id, role)
		);

		ALTER TABLE accounts ADD COLUMN IF NOT EXISTS class_id TEXT REFERENCES school_classes(id);
		ALTER TABLE registration_requests ADD COLUMN IF NOT EXISTS class_ids TEXT[] NOT NULL DEFAULT '{}';
		ALTER TABLE user_profiles ADD COLUMN IF NOT EXISTS class_ids TEXT[] NOT NULL DEFAULT '{}';

		-- Indexes for new user management tables
		CREATE INDEX IF NOT EXISTS idx_user_registry_person_id ON user_registry(person_id);
		CREATE INDEX IF NOT EXISTS idx_user_registry_status ON user_registry(registration_status);
		CREATE INDEX IF NOT EXISTS idx_user_registry_email ON user_registry(email);

		CREATE INDEX IF NOT EXISTS idx_registration_workflow_user_id ON registration_workflow(user_id);
		CREATE INDEX IF NOT EXISTS idx_registration_workflow_status ON registration_workflow(approval_status);
		CREATE INDEX IF NOT EXISTS idx_registration_workflow_created_at ON registration_workflow(created_at);

		-- Initialize role mappings with default values
		INSERT INTO role_mappings (keycloak_role, local_role, description, auto_assign, requires_approval, can_teach, can_manage_classes, can_manage_users, can_view_reports, show_in_registration, registration_order)
		VALUES 
		    ('teacher', 'teacher', 'Teaching staff member', TRUE, FALSE, TRUE, FALSE, FALSE, FALSE, TRUE, 1),
		    ('student', 'student', 'Student at the school', TRUE, FALSE, FALSE, FALSE, FALSE, FALSE, TRUE, 2),
		    ('guardian', 'guardian', 'Parent or legal guardian', TRUE, FALSE, FALSE, FALSE, FALSE, FALSE, TRUE, 3),
		    ('school-management', 'school_management', 'School administrator', FALSE, TRUE, FALSE, TRUE, FALSE, TRUE, FALSE, 4),
		    ('admin', 'admin', 'Administrator with full access', FALSE, FALSE, FALSE, TRUE, TRUE, TRUE, FALSE, 5)
		ON CONFLICT (keycloak_role) DO NOTHING;

		CREATE TABLE IF NOT EXISTS receipt_splits (
			id         SERIAL        PRIMARY KEY,
			receipt_id TEXT          NOT NULL REFERENCES receipts(id) ON DELETE CASCADE,
			class_id   TEXT          REFERENCES school_classes(id),
			account_id TEXT          REFERENCES accounts(id),
			amount     NUMERIC(12,2) NOT NULL
		);

		-- ── Step 3: extend existing tables + new tables that depend on them ───────

		ALTER TABLE persons ADD COLUMN IF NOT EXISTS prename       TEXT NOT NULL DEFAULT '';
		ALTER TABLE persons ADD COLUMN IF NOT EXISTS date_of_birth TIMESTAMPTZ;
		ALTER TABLE persons ADD COLUMN IF NOT EXISTS address_id    TEXT REFERENCES addresses(id);
		ALTER TABLE persons ADD COLUMN IF NOT EXISTS version       INT NOT NULL DEFAULT 0;

		ALTER TABLE school_classes ALTER COLUMN name DROP NOT NULL;
		ALTER TABLE school_classes ADD COLUMN IF NOT EXISTS shortcut       TEXT;
		ALTER TABLE school_classes ADD COLUMN IF NOT EXISTS classroom_id   TEXT REFERENCES rooms(id);
		ALTER TABLE school_classes ADD COLUMN IF NOT EXISTS school_year_id TEXT REFERENCES school_years(id);
		ALTER TABLE school_classes ADD COLUMN IF NOT EXISTS version        INT NOT NULL DEFAULT 0;

		-- Migrate postal_codes from number PK to UUID id PK
		-- First check and drop all FKs that reference the old postal_codes(number) PK
		DO $$ BEGIN
		    PERFORM 1 FROM pg_constraint WHERE conname = 'city_postal_codes_postal_code_no_fkey';
		    IF FOUND THEN
		        EXECUTE 'ALTER TABLE city_postal_codes DROP CONSTRAINT city_postal_codes_postal_code_no_fkey CASCADE';
		    END IF;
		EXCEPTION WHEN OTHERS THEN NULL;
		END $$;

		DO $$ BEGIN
		    PERFORM 1 FROM pg_constraint WHERE conname = 'addresses_postal_code_no_fkey';
		    IF FOUND THEN
		        EXECUTE 'ALTER TABLE addresses DROP CONSTRAINT addresses_postal_code_no_fkey CASCADE';
		    END IF;
		EXCEPTION WHEN OTHERS THEN NULL;
		END $$;

		-- Now safe to drop the old PK and add the new one
		ALTER TABLE postal_codes ADD COLUMN IF NOT EXISTS id TEXT;
		UPDATE postal_codes SET id = gen_random_uuid()::text WHERE id IS NULL;
		ALTER TABLE postal_codes ALTER COLUMN id SET NOT NULL;
		
		DO $$ BEGIN
		    PERFORM 1 FROM pg_constraint WHERE conname = 'postal_codes_pkey';
		    IF FOUND THEN
		        EXECUTE 'ALTER TABLE postal_codes DROP CONSTRAINT postal_codes_pkey CASCADE';
		    END IF;
		EXCEPTION WHEN OTHERS THEN NULL;
		END $$;

		ALTER TABLE postal_codes ADD PRIMARY KEY (id);
		ALTER TABLE postal_codes ALTER COLUMN number DROP NOT NULL;

		-- Migrate city_postal_codes FK: postal_code_no → postal_code_id
		ALTER TABLE city_postal_codes ADD COLUMN IF NOT EXISTS postal_code_id TEXT;
		
		-- Only update if postal_code_no column exists
		DO $$ BEGIN
		    IF EXISTS (SELECT 1 FROM information_schema.columns 
		                WHERE table_name = 'city_postal_codes' AND column_name = 'postal_code_no') THEN
		        UPDATE city_postal_codes cpc SET postal_code_id = pc.id
		            FROM postal_codes pc WHERE pc.number = cpc.postal_code_no;
		    END IF;
		EXCEPTION WHEN OTHERS THEN NULL;
		END $$;
		
		ALTER TABLE city_postal_codes DROP CONSTRAINT IF EXISTS city_postal_codes_pkey;
		ALTER TABLE city_postal_codes DROP COLUMN IF EXISTS postal_code_no;
		ALTER TABLE city_postal_codes ADD PRIMARY KEY (city_id, postal_code_id);
		ALTER TABLE city_postal_codes ADD CONSTRAINT city_postal_codes_postal_code_id_fkey
			FOREIGN KEY (postal_code_id) REFERENCES postal_codes(id) ON DELETE CASCADE;

		-- Migrate addresses FK: postal_code_no → postal_code_id
		ALTER TABLE addresses ADD COLUMN IF NOT EXISTS postal_code_id TEXT;
		
		-- Only update if postal_code_no column exists
		DO $$ BEGIN
		    IF EXISTS (SELECT 1 FROM information_schema.columns 
		                WHERE table_name = 'addresses' AND column_name = 'postal_code_no') THEN
		        UPDATE addresses a SET postal_code_id = pc.id
		            FROM postal_codes pc WHERE pc.number = a.postal_code_no;
		    END IF;
		EXCEPTION WHEN OTHERS THEN NULL;
		END $$;
		
		ALTER TABLE addresses DROP COLUMN IF EXISTS postal_code_no;
		ALTER TABLE addresses ADD CONSTRAINT addresses_postal_code_id_fkey
			FOREIGN KEY (postal_code_id) REFERENCES postal_codes(id);

		CREATE TABLE IF NOT EXISTS students (
			person_id               TEXT PRIMARY KEY REFERENCES persons(id) ON DELETE CASCADE,
			current_school_class_id TEXT REFERENCES school_classes(id)
		);

		CREATE TABLE IF NOT EXISTS student_guardians (
			student_id  TEXT NOT NULL REFERENCES students(person_id) ON DELETE CASCADE,
			guardian_id TEXT NOT NULL REFERENCES guardians(person_id) ON DELETE CASCADE,
			PRIMARY KEY (student_id, guardian_id)
		);

		CREATE TABLE IF NOT EXISTS student_past_classes (
			student_id      TEXT NOT NULL REFERENCES students(person_id) ON DELETE CASCADE,
			school_class_id TEXT NOT NULL REFERENCES school_classes(id) ON DELETE CASCADE,
			PRIMARY KEY (student_id, school_class_id)
		);

		CREATE TABLE IF NOT EXISTS school_class_teachers (
			school_class_id TEXT NOT NULL REFERENCES school_classes(id) ON DELETE CASCADE,
			teacher_id      TEXT NOT NULL REFERENCES teachers(person_id) ON DELETE CASCADE,
			PRIMARY KEY (school_class_id, teacher_id)
		);

		CREATE TABLE IF NOT EXISTS school_class_students (
			school_class_id TEXT NOT NULL REFERENCES school_classes(id) ON DELETE CASCADE,
			student_id      TEXT NOT NULL REFERENCES students(person_id) ON DELETE CASCADE,
			PRIMARY KEY (school_class_id, student_id)
		);

		CREATE TABLE IF NOT EXISTS subject_teachers (
			subject_id TEXT NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
			teacher_id TEXT NOT NULL REFERENCES teachers(person_id) ON DELETE CASCADE,
			PRIMARY KEY (subject_id, teacher_id)
		);

		CREATE TABLE IF NOT EXISTS lessons (
			id              TEXT PRIMARY KEY,
			teacher_id      TEXT REFERENCES teachers(person_id),
			school_class_id TEXT REFERENCES school_classes(id),
			subject_id      TEXT REFERENCES subjects(id),
			version         INT NOT NULL DEFAULT 0
		);
		ALTER TABLE rooms ADD COLUMN IF NOT EXISTS room_type TEXT;

		ALTER TABLE lessons ADD COLUMN IF NOT EXISTS day_of_week INT;
		ALTER TABLE lessons ADD COLUMN IF NOT EXISTS period INT;
		ALTER TABLE lessons ADD COLUMN IF NOT EXISTS start_time TEXT;
		ALTER TABLE lessons ADD COLUMN IF NOT EXISTS end_time TEXT;
		ALTER TABLE lessons ADD COLUMN IF NOT EXISTS room_id TEXT REFERENCES rooms(id);

		CREATE TABLE IF NOT EXISTS exams (
			id              TEXT PRIMARY KEY,
			school_class_id TEXT REFERENCES school_classes(id),
			version         INT NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS grades (
			id         TEXT PRIMARY KEY,
			grade      NUMERIC(5,2),
			exam_id    TEXT REFERENCES exams(id),
			student_id TEXT REFERENCES students(person_id),
			version    INT NOT NULL DEFAULT 0
		);

		-- Drop FK constraints on audit columns that store Keycloak subs, not user_registry ids.
		ALTER TABLE registration_workflow DROP CONSTRAINT IF EXISTS registration_workflow_approval_by_fkey;
		ALTER TABLE registration_workflow DROP CONSTRAINT IF EXISTS registration_workflow_rejection_by_fkey;
	`)
	return err
}

// — role mappings —

func (p *Postgres) CreateRoleMapping(ctx context.Context, mapping *model.RoleMapping) error {
	mapping.CreatedAt = time.Now()
	mapping.UpdatedAt = time.Now()

	_, err := p.db.ExecContext(ctx,
		`INSERT INTO role_mappings
		 (keycloak_role, local_role, description, auto_assign, auto_create,
		  requires_approval, can_teach, can_manage_classes, can_manage_users,
		  can_view_reports, show_in_registration, registration_order,
		  created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		mapping.KeycloakRole, mapping.LocalRole, mapping.Description, mapping.AutoAssign,
		mapping.AutoCreate, mapping.RequiresApproval, mapping.CanTeach, mapping.CanManageClasses,
		mapping.CanManageUsers, mapping.CanViewReports, mapping.ShowInRegistration,
		mapping.RegistrationOrder, mapping.CreatedAt, mapping.UpdatedAt,
	)
	return err
}

func (p *Postgres) GetRoleMapping(ctx context.Context, keycloakRole string) (*model.RoleMapping, error) {
	mapping := &model.RoleMapping{}

	err := p.db.QueryRowContext(ctx,
		`SELECT keycloak_role, local_role, description, auto_assign, auto_create,
		 requires_approval, can_teach, can_manage_classes, can_manage_users,
		 can_view_reports, show_in_registration, registration_order,
		 created_at, updated_at
		 FROM role_mappings WHERE keycloak_role = $1`, keycloakRole,
	).Scan(&mapping.KeycloakRole, &mapping.LocalRole, &mapping.Description,
		&mapping.AutoAssign, &mapping.AutoCreate, &mapping.RequiresApproval,
		&mapping.CanTeach, &mapping.CanManageClasses, &mapping.CanManageUsers,
		&mapping.CanViewReports, &mapping.ShowInRegistration, &mapping.RegistrationOrder,
		&mapping.CreatedAt, &mapping.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting role mapping: %w", err)
	}

	return mapping, nil
}

func (p *Postgres) ListRoleMappings(ctx context.Context) ([]*model.RoleMapping, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT keycloak_role, local_role, description, auto_assign, auto_create,
		 requires_approval, can_teach, can_manage_classes, can_manage_users,
		 can_view_reports, show_in_registration, registration_order,
		 created_at, updated_at
		 FROM role_mappings ORDER BY registration_order`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing role mappings: %w", err)
	}
	defer rows.Close()

	var mappings []*model.RoleMapping
	for rows.Next() {
		mapping := &model.RoleMapping{}

		err := rows.Scan(&mapping.KeycloakRole, &mapping.LocalRole, &mapping.Description,
			&mapping.AutoAssign, &mapping.AutoCreate, &mapping.RequiresApproval,
			&mapping.CanTeach, &mapping.CanManageClasses, &mapping.CanManageUsers,
			&mapping.CanViewReports, &mapping.ShowInRegistration, &mapping.RegistrationOrder,
			&mapping.CreatedAt, &mapping.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning role mapping: %w", err)
		}

		mappings = append(mappings, mapping)
	}

	return mappings, nil
}

func (p *Postgres) UpdateRoleMapping(ctx context.Context, mapping *model.RoleMapping) error {
	mapping.UpdatedAt = time.Now()

	_, err := p.db.ExecContext(ctx,
		`UPDATE role_mappings SET
		 local_role = $1, description = $2, auto_assign = $3, auto_create = $4,
		 requires_approval = $5, can_teach = $6, can_manage_classes = $7,
		 can_manage_users = $8, can_view_reports = $9, show_in_registration = $10,
		 registration_order = $11, updated_at = $12
		 WHERE keycloak_role = $13`,
		mapping.LocalRole, mapping.Description, mapping.AutoAssign, mapping.AutoCreate,
		mapping.RequiresApproval, mapping.CanTeach, mapping.CanManageClasses,
		mapping.CanManageUsers, mapping.CanViewReports, mapping.ShowInRegistration,
		mapping.RegistrationOrder, mapping.UpdatedAt, mapping.KeycloakRole,
	)
	return err
}

func (p *Postgres) DeleteRoleMapping(ctx context.Context, keycloakRole string) error {
	result, err := p.db.ExecContext(ctx,
		`DELETE FROM role_mappings WHERE keycloak_role = $1`,
		keycloakRole,
	)
	if err != nil {
		return fmt.Errorf("deleting role mapping: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("role mapping not found")
	}

	return nil
}

// — user linking —

func (p *Postgres) EnsureUserLinked(ctx context.Context, userSub string, personData *model.Person) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Get or create user registry entry
	user, err := p.getOrCreateUserRegistry(ctx, tx, userSub)
	if err != nil {
		return fmt.Errorf("creating user registry: %w", err)
	}

	// 2. Get or create person record
	person, err := p.getOrCreatePerson(ctx, tx, userSub, personData)
	if err != nil {
		return fmt.Errorf("creating person: %w", err)
	}

	// 3. Link user to person if not already linked
	if user.PersonID != person.ID {
		if _, err := tx.ExecContext(ctx,
			`UPDATE user_registry SET person_id = $1 WHERE id = $2`,
			person.ID, user.ID,
		); err != nil {
			return fmt.Errorf("linking user to person: %w", err)
		}
	}

	// 4. Sync roles from Keycloak
	if err := p.syncRolesFromKeycloak(ctx, tx, userSub, user.ID); err != nil {
		return fmt.Errorf("syncing roles: %w", err)
	}

	return tx.Commit()
}

func (p *Postgres) SyncUserRoles(ctx context.Context, userSub string) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get user by sub
	user, err := p.GetUserRegistryBySub(ctx, userSub)
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Get current Keycloak roles from JWT (this would be passed in context in real implementation)
	// For now, we'll assume we get them from the user's current keycloak_roles
	// In a real implementation, this would come from the JWT claims
	currentKeycloakRoles := user.KeycloakRoles

	// Get role mappings
	mappings, err := p.ListRoleMappings(ctx)
	if err != nil {
		return fmt.Errorf("getting role mappings: %w", err)
	}

	// Build new local roles based on Keycloak roles and mappings
	var newLocalRoles []string
	for _, kcRole := range currentKeycloakRoles {
		for _, mapping := range mappings {
			if mapping.KeycloakRole == kcRole && mapping.AutoAssign {
				newLocalRoles = append(newLocalRoles, mapping.LocalRole)
				break
			}
		}
	}

	// Only update if roles have changed
	if !slicesEqual(user.LocalRoles, newLocalRoles) {
		user.LocalRoles = newLocalRoles
		user.LastRoleSync = ptr(time.Now())

		if err := p.UpdateUserRegistry(ctx, user); err != nil {
			return fmt.Errorf("updating user roles: %w", err)
		}
	}

	return tx.Commit()
}

// Helper function to get or create user registry entry
func (p *Postgres) getOrCreateUserRegistry(ctx context.Context, tx *sql.Tx, userSub string) (*model.UserRegistry, error) {
	// Check if user exists
	user, err := p.getUserRegistryBySub(ctx, tx, userSub)
	if err != nil {
		return nil, err
	}

	if user != nil {
		return user, nil
	}

	// Create new user registry entry
	user = &model.UserRegistry{
		UserSub:            userSub,
		RegistrationStatus: model.RegistrationStatusPending,
	}

	if err := p.createUserRegistry(ctx, tx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Helper function to get user registry by sub (within transaction)
func (p *Postgres) getUserRegistryBySub(ctx context.Context, tx *sql.Tx, userSub string) (*model.UserRegistry, error) {
	user := &model.UserRegistry{}
	var personID, email, createdBy, updatedBy, legacyProfileID, legacyRequestID sql.NullString
	var lastRoleSync, lastLoginAt sql.NullTime

	err := tx.QueryRowContext(ctx,
		`SELECT id, user_sub, person_id, email, registration_status, keycloak_roles, local_roles,
		 last_role_sync, created_at, updated_at, last_login_at, created_by, updated_by,
		 legacy_profile_id, legacy_request_id
		 FROM user_registry WHERE user_sub = $1`, userSub,
	).Scan(&user.ID, &user.UserSub, &personID, &email, &user.RegistrationStatus,
		pq.Array(&user.KeycloakRoles), pq.Array(&user.LocalRoles),
		&lastRoleSync, &user.CreatedAt, &user.UpdatedAt, &lastLoginAt,
		&createdBy, &updatedBy, &legacyProfileID, &legacyRequestID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting user registry by sub: %w", err)
	}

	user.PersonID = personID.String
	user.Email = email.String
	if lastRoleSync.Valid {
		t := time.Time(lastRoleSync.Time)
		user.LastRoleSync = &t
	}
	if lastLoginAt.Valid {
		t := time.Time(lastLoginAt.Time)
		user.LastLoginAt = &t
	}
	user.CreatedBy = createdBy.String
	user.UpdatedBy = updatedBy.String
	user.LegacyProfileID = legacyProfileID.String
	user.LegacyRequestID = legacyRequestID.String

	return user, nil
}

// Helper function to create user registry entry (within transaction)
func (p *Postgres) createUserRegistry(ctx context.Context, tx *sql.Tx, user *model.UserRegistry) error {
	user.ID = newID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := tx.ExecContext(ctx,
		`INSERT INTO user_registry 
		 (id, user_sub, registration_status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		user.ID, user.UserSub, user.RegistrationStatus, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

// Helper function to get or create person record
func (p *Postgres) getOrCreatePerson(ctx context.Context, tx *sql.Tx, userSub string, personData *model.Person) (*model.Person, error) {
	if personData == nil {
		// Check if person exists
		person, err := p.getPersonBySub(ctx, tx, userSub)
		if err != nil {
			return nil, err
		}
		if person != nil {
			return person, nil
		}
		return nil, fmt.Errorf("person data required for new person")
	}

	// Check if person exists
	person, err := p.getPersonBySub(ctx, tx, userSub)
	if err != nil {
		return nil, err
	}

	if person != nil {
		// Update existing person if needed
		person.Name = personData.Name
		person.Prename = personData.Prename
		if personData.DateOfBirth != nil {
			person.DateOfBirth = personData.DateOfBirth
		}
		if personData.Address != nil {
			person.Address = personData.Address
		}

		if err := p.updatePerson(ctx, tx, person); err != nil {
			return nil, err
		}
		return person, nil
	}

	// Create new person
	person = personData
	person.ID = newID()
	person.Sub = &userSub
	person.Version = 0

	if err := p.createPerson(ctx, tx, person); err != nil {
		return nil, err
	}

	return person, nil
}

// Helper function to get person by sub (within transaction)
func (p *Postgres) getPersonBySub(ctx context.Context, tx *sql.Tx, userSub string) (*model.Person, error) {
	person := &model.Person{}
	var name, prename sql.NullString
	var dateOfBirth sql.NullTime
	var addressID sql.NullString

	var subStr sql.NullString
	err := tx.QueryRowContext(ctx,
		`SELECT id, name, prename, date_of_birth, address_id, sub, version
		 FROM persons WHERE sub = $1`, userSub,
	).Scan(&person.ID, &name, &prename, &dateOfBirth, &addressID, &subStr, &person.Version)
	if subStr.Valid && subStr.String != "" {
		person.Sub = &subStr.String
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting person by sub: %w", err)
	}

	person.Name = name.String
	person.Prename = prename.String
	if dateOfBirth.Valid {
		t := time.Time(dateOfBirth.Time)
		person.DateOfBirth = &t
	}
	if addressID.Valid {
		// Note: We don't load the full address object here for simplicity
		person.Address = &model.Address{ID: addressID.String}
	}

	return person, nil
}

// Helper function to create person (within transaction)
func (p *Postgres) createPerson(ctx context.Context, tx *sql.Tx, person *model.Person) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO persons 
		 (id, name, prename, date_of_birth, address_id, sub, version)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		person.ID, person.Name, person.Prename, nullableTime(person.DateOfBirth),
		nullableString(getAddressID(person.Address)), nullableStringPtr(person.Sub), person.Version,
	)
	return err
}

// Helper function to update person (within transaction)
func (p *Postgres) updatePerson(ctx context.Context, tx *sql.Tx, person *model.Person) error {
	person.Version++

	_, err := tx.ExecContext(ctx,
		`UPDATE persons SET
		 name = $1, prename = $2, date_of_birth = $3, address_id = $4, version = $5
		 WHERE id = $6`,
		person.Name, person.Prename, nullableTime(person.DateOfBirth),
		nullableString(getAddressID(person.Address)), person.Version, person.ID,
	)
	return err
}

// Helper function to sync roles from Keycloak
func (p *Postgres) syncRolesFromKeycloak(ctx context.Context, tx *sql.Tx, userSub string, userID string) error {
	// In a real implementation, this would:
	// 1. Get current Keycloak roles from the JWT claims
	// 2. Compare with current local roles
	// 3. Update local roles based on role mappings
	// 4. Potentially create role-specific entities (teacher, student, etc.)

	// For now, this is a placeholder implementation
	// The actual implementation would depend on how Keycloak roles are passed to the backend

	return nil
}

// Helper function to compare string slices
func slicesEqual(a, b []string) bool {
	return slices.Equal(a, b)
}

// Helper function to create time pointer
func ptr(t time.Time) *time.Time {
	return &t
}

// Create inserts a new receipt and its items. The receipt ID is generated here.
// Status is always forced to unsubmitted regardless of what the caller sets.
func (p *Postgres) Create(ctx context.Context, r *model.Receipt) error {
	r.ID = newID()
	r.Status = model.Unsubmitted

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO receipts (id, file_id, receipt_owner_id, total_price, taxes, time, account_id, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		r.ID, nullableString(r.FileID), r.ReceiptOwnerID, r.TotalPrice, r.Taxes, r.Time, nullableString(r.AccountID), r.Status,
	)
	if err != nil {
		return fmt.Errorf("inserting receipt: %w", err)
	}

	if err := insertItems(ctx, tx, r.ID, r.Items); err != nil {
		return err
	}
	if err := insertSplits(ctx, tx, r.ID, r.Splits); err != nil {
		return err
	}

	return tx.Commit()
}

// Get returns a single receipt by ID, including its items and splits.
func (p *Postgres) Get(ctx context.Context, id string) (*model.Receipt, error) {
	r := &model.Receipt{}
	var accountID, fileID sql.NullString
	err := p.db.QueryRowContext(ctx,
		`SELECT id, file_id, receipt_owner_id, total_price, taxes, time, account_id, status FROM receipts WHERE id = $1`, id,
	).Scan(&r.ID, &fileID, &r.ReceiptOwnerID, &r.TotalPrice, &r.Taxes, &r.Time, &accountID, &r.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying receipt: %w", err)
	}
	r.AccountID = accountID.String
	r.FileID = fileID.String

	items, err := p.queryItems(ctx, id)
	if err != nil {
		return nil, err
	}
	r.Items = items

	splits, err := p.querySplits(ctx, id)
	if err != nil {
		return nil, err
	}
	r.Splits = splits
	return r, nil
}

// List returns all receipts. If ownerID is non-empty, filters by receipt_owner_id.
func (p *Postgres) List(ctx context.Context, ownerID string) ([]*model.Receipt, error) {
	query := `SELECT id, file_id, receipt_owner_id, total_price, taxes, time, account_id, status FROM receipts`
	args := []any{}
	if ownerID != "" {
		query += ` WHERE receipt_owner_id = $1`
		args = append(args, ownerID)
	}
	query += ` ORDER BY time DESC`

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing receipts: %w", err)
	}
	defer rows.Close()

	var receipts []*model.Receipt
	for rows.Next() {
		r := &model.Receipt{}
		var accountID, fileID sql.NullString
		if err := rows.Scan(&r.ID, &fileID, &r.ReceiptOwnerID, &r.TotalPrice, &r.Taxes, &r.Time, &accountID, &r.Status); err != nil {
			return nil, err
		}
		r.AccountID = accountID.String
		r.FileID = fileID.String
		receipts = append(receipts, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(receipts) > 0 {
		if err := p.attachItems(ctx, receipts); err != nil {
			return nil, err
		}
		if err := p.attachSplits(ctx, receipts); err != nil {
			return nil, err
		}
	}

	return receipts, nil
}

// Update replaces a receipt's data, all its items, and all its splits.
func (p *Postgres) Update(ctx context.Context, r *model.Receipt) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE receipts SET file_id=$1, receipt_owner_id=$2, total_price=$3, taxes=$4, time=$5, account_id=$6, status=$7 WHERE id=$8`,
		nullableString(r.FileID), r.ReceiptOwnerID, r.TotalPrice, r.Taxes, r.Time, nullableString(r.AccountID), r.Status, r.ID,
	)
	if err != nil {
		return fmt.Errorf("updating receipt: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("receipt not found")
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM receipt_items WHERE receipt_id = $1`, r.ID); err != nil {
		return fmt.Errorf("clearing items: %w", err)
	}
	if err := insertItems(ctx, tx, r.ID, r.Items); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM receipt_splits WHERE receipt_id = $1`, r.ID); err != nil {
		return fmt.Errorf("clearing splits: %w", err)
	}
	if err := insertSplits(ctx, tx, r.ID, r.Splits); err != nil {
		return err
	}

	return tx.Commit()
}

// Delete removes a receipt and its items (via CASCADE).
func (p *Postgres) Delete(ctx context.Context, id string) error {
	res, err := p.db.ExecContext(ctx, `DELETE FROM receipts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting receipt: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("receipt not found")
	}
	return nil
}

// UpdateStatus sets the status field for multiple receipts in one query.
func (p *Postgres) UpdateStatus(ctx context.Context, ids []string, status model.ReceiptStatus) error {
	if len(ids) == 0 {
		return nil
	}
	args := make([]any, len(ids)+1)
	args[0] = string(status)
	placeholders := make([]byte, 0, len(ids)*4)
	for i, id := range ids {
		if i > 0 {
			placeholders = append(placeholders, ',')
		}
		placeholders = fmt.Appendf(placeholders, "$%d", i+2)
		args[i+1] = id
	}
	_, err := p.db.ExecContext(ctx,
		`UPDATE receipts SET status = $1 WHERE id IN (`+string(placeholders)+`)`,
		args...,
	)
	return err
}

// — helpers —

func insertItems(ctx context.Context, tx *sql.Tx, receiptID string, items []model.ReceiptItem) error {
	for _, it := range items {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO receipt_items (receipt_id, name, amount, price) VALUES ($1, $2, $3, $4)`,
			receiptID, it.Name, it.Amount, it.Price,
		)
		if err != nil {
			return fmt.Errorf("inserting item: %w", err)
		}
	}
	return nil
}

func (p *Postgres) queryItems(ctx context.Context, receiptID string) ([]model.ReceiptItem, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT name, amount, price FROM receipt_items WHERE receipt_id = $1 ORDER BY id`, receiptID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying items: %w", err)
	}
	defer rows.Close()

	var items []model.ReceiptItem
	for rows.Next() {
		var it model.ReceiptItem
		if err := rows.Scan(&it.Name, &it.Amount, &it.Price); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// attachItems loads items for multiple receipts in one DB round-trip.
func (p *Postgres) attachItems(ctx context.Context, receipts []*model.Receipt) error {
	ids := make([]any, len(receipts))
	index := make(map[string]*model.Receipt, len(receipts))
	for i, r := range receipts {
		ids[i] = r.ID
		index[r.ID] = r
	}

	// Build $1,$2,... placeholder list.
	placeholders := make([]byte, 0, len(ids)*4)
	for i := range ids {
		if i > 0 {
			placeholders = append(placeholders, ',')
		}
		placeholders = fmt.Appendf(placeholders, "$%d", i+1)
	}

	rows, err := p.db.QueryContext(ctx,
		`SELECT receipt_id, name, amount, price FROM receipt_items WHERE receipt_id IN (`+string(placeholders)+`) ORDER BY id`,
		ids...,
	)
	if err != nil {
		return fmt.Errorf("querying items in bulk: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var receiptID string
		var it model.ReceiptItem
		if err := rows.Scan(&receiptID, &it.Name, &it.Amount, &it.Price); err != nil {
			return err
		}
		r := index[receiptID]
		r.Items = append(r.Items, it)
	}
	return rows.Err()
}

// — split helpers —

func insertSplits(ctx context.Context, tx *sql.Tx, receiptID string, splits []model.ReceiptSplit) error {
	for _, s := range splits {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO receipt_splits (receipt_id, class_id, account_id, amount) VALUES ($1, $2, $3, $4)`,
			receiptID, nullableString(s.ClassID), nullableString(s.AccountID), s.Amount,
		)
		if err != nil {
			return fmt.Errorf("inserting split: %w", err)
		}
	}
	return nil
}

func (p *Postgres) querySplits(ctx context.Context, receiptID string) ([]model.ReceiptSplit, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT COALESCE(class_id,''), COALESCE(account_id,''), amount FROM receipt_splits WHERE receipt_id = $1 ORDER BY id`,
		receiptID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying splits: %w", err)
	}
	defer rows.Close()
	var splits []model.ReceiptSplit
	for rows.Next() {
		var s model.ReceiptSplit
		if err := rows.Scan(&s.ClassID, &s.AccountID, &s.Amount); err != nil {
			return nil, err
		}
		splits = append(splits, s)
	}
	return splits, rows.Err()
}

func (p *Postgres) attachSplits(ctx context.Context, receipts []*model.Receipt) error {
	ids := make([]any, len(receipts))
	index := make(map[string]*model.Receipt, len(receipts))
	for i, r := range receipts {
		ids[i] = r.ID
		index[r.ID] = r
	}
	placeholders := make([]byte, 0, len(ids)*4)
	for i := range ids {
		if i > 0 {
			placeholders = append(placeholders, ',')
		}
		placeholders = fmt.Appendf(placeholders, "$%d", i+1)
	}
	rows, err := p.db.QueryContext(ctx,
		`SELECT receipt_id, COALESCE(class_id,''), COALESCE(account_id,''), amount FROM receipt_splits WHERE receipt_id IN (`+string(placeholders)+`) ORDER BY id`,
		ids...,
	)
	if err != nil {
		return fmt.Errorf("querying splits in bulk: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var receiptID string
		var s model.ReceiptSplit
		if err := rows.Scan(&receiptID, &s.ClassID, &s.AccountID, &s.Amount); err != nil {
			return err
		}
		r := index[receiptID]
		r.Splits = append(r.Splits, s)
	}
	return rows.Err()
}

// — account CRUD —

func (p *Postgres) CreateAccount(ctx context.Context, a *model.Account) error {
	a.ID = newID()
	now := time.Now()
	if a.ValidFrom.IsZero() {
		a.ValidFrom = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if a.ValidTo.IsZero() {
		a.ValidTo = time.Date(now.Year(), 12, 31, 0, 0, 0, 0, time.UTC)
	}
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO accounts (id, name, shortcut, budget, class_id, valid_from, valid_to) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		a.ID, a.Name, a.Shortcut, a.Budget, nullableString(a.ClassID), a.ValidFrom, a.ValidTo,
	)
	return err
}

func (p *Postgres) GetAccount(ctx context.Context, id string) (*model.Account, error) {
	a := &model.Account{}
	var classID sql.NullString
	err := p.db.QueryRowContext(ctx, `
		SELECT a.id, a.name, a.shortcut, a.budget, COALESCE(a.class_id,''), a.valid_from, a.valid_to,
		       COALESCE(SUM(r.total_price), 0) AS spent
		FROM accounts a
		LEFT JOIN receipts r ON r.account_id = a.id
		WHERE a.id = $1
		GROUP BY a.id, a.name, a.shortcut, a.budget, a.class_id, a.valid_from, a.valid_to`, id,
	).Scan(&a.ID, &a.Name, &a.Shortcut, &a.Budget, &classID, &a.ValidFrom, &a.ValidTo, &a.Spent)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying account: %w", err)
	}
	a.ClassID = classID.String
	a.Balance = a.Budget - a.Spent
	return a, nil
}

func (p *Postgres) ListAccounts(ctx context.Context) ([]*model.Account, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT a.id, a.name, a.shortcut, a.budget, COALESCE(a.class_id,''), a.valid_from, a.valid_to,
		       COALESCE(SUM(r.total_price), 0) AS spent
		FROM accounts a
		LEFT JOIN receipts r ON r.account_id = a.id
		GROUP BY a.id, a.name, a.shortcut, a.budget, a.class_id, a.valid_from, a.valid_to
		ORDER BY a.name`)
	if err != nil {
		return nil, fmt.Errorf("listing accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*model.Account
	for rows.Next() {
		a := &model.Account{}
		var classID sql.NullString
		if err := rows.Scan(&a.ID, &a.Name, &a.Shortcut, &a.Budget, &classID, &a.ValidFrom, &a.ValidTo, &a.Spent); err != nil {
			return nil, err
		}
		a.ClassID = classID.String
		a.Balance = a.Budget - a.Spent
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

func (p *Postgres) UpdateAccount(ctx context.Context, a *model.Account) error {
	res, err := p.db.ExecContext(ctx,
		`UPDATE accounts SET name=$1, shortcut=$2, budget=$3, class_id=$4, valid_from=$5, valid_to=$6 WHERE id=$7`,
		a.Name, a.Shortcut, a.Budget, nullableString(a.ClassID), a.ValidFrom, a.ValidTo, a.ID,
	)
	if err != nil {
		return fmt.Errorf("updating account: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("account not found")
	}
	return nil
}

func (p *Postgres) DeleteAccount(ctx context.Context, id string) error {
	res, err := p.db.ExecContext(ctx, `DELETE FROM accounts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting account: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("account not found")
	}
	return nil
}

// — registration requests —

func (p *Postgres) CreateRegistrationRequest(ctx context.Context, req *model.RegistrationRequest) error {
	req.ID = newID()
	if req.ClassIDs == nil {
		req.ClassIDs = []string{}
	}
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO registration_requests (id, user_sub, email, desired_roles, class_ids, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		req.ID, req.UserSub, req.Email, pq.Array(req.DesiredRoles), pq.Array(req.ClassIDs), string(req.Status), req.CreatedAt,
	)
	return err
}

func (p *Postgres) GetRegistrationRequestByUserSub(ctx context.Context, userSub string) (*model.RegistrationRequest, error) {
	req := &model.RegistrationRequest{}
	err := p.db.QueryRowContext(ctx,
		`SELECT id, user_sub, email, desired_roles, class_ids, status, created_at
		 FROM registration_requests WHERE user_sub = $1`, userSub,
	).Scan(&req.ID, &req.UserSub, &req.Email, pq.Array(&req.DesiredRoles), pq.Array(&req.ClassIDs), &req.Status, &req.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying registration request: %w", err)
	}
	return req, nil
}

func (p *Postgres) GetRegistrationRequestByID(ctx context.Context, id string) (*model.RegistrationRequest, error) {
	req := &model.RegistrationRequest{}
	err := p.db.QueryRowContext(ctx,
		`SELECT id, user_sub, email, desired_roles, class_ids, status, created_at
		 FROM registration_requests WHERE id = $1`, id,
	).Scan(&req.ID, &req.UserSub, &req.Email, pq.Array(&req.DesiredRoles), pq.Array(&req.ClassIDs), &req.Status, &req.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying registration request by id: %w", err)
	}
	return req, nil
}

func (p *Postgres) ListRegistrationRequests(ctx context.Context) ([]*model.RegistrationRequest, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT id, user_sub, email, desired_roles, class_ids, status, created_at
		 FROM registration_requests ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("listing registration requests: %w", err)
	}
	defer rows.Close()

	var reqs []*model.RegistrationRequest
	for rows.Next() {
		req := &model.RegistrationRequest{}
		if err := rows.Scan(&req.ID, &req.UserSub, &req.Email, pq.Array(&req.DesiredRoles), pq.Array(&req.ClassIDs), &req.Status, &req.CreatedAt); err != nil {
			return nil, err
		}
		reqs = append(reqs, req)
	}
	return reqs, rows.Err()
}

func (p *Postgres) UpdateRegistrationRequestStatus(ctx context.Context, id string, status model.RegistrationStatus) error {
	res, err := p.db.ExecContext(ctx,
		`UPDATE registration_requests SET status = $1 WHERE id = $2`,
		string(status), id,
	)
	if err != nil {
		return fmt.Errorf("updating registration request status: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("registration request not found")
	}
	return nil
}

// — user profiles —

func (p *Postgres) GetUserProfile(ctx context.Context, userSub string) (*model.UserProfile, error) {
	profile := &model.UserProfile{}
	err := p.db.QueryRowContext(ctx,
		`SELECT user_sub, iban, address, phone, class_ids, profile_complete, profile_skipped
		 FROM user_profiles WHERE user_sub = $1`, userSub,
	).Scan(&profile.UserSub, &profile.IBAN, &profile.Address, &profile.Phone, pq.Array(&profile.ClassIDs), &profile.ProfileComplete, &profile.ProfileSkipped)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying user profile: %w", err)
	}
	return profile, nil
}

func (p *Postgres) UpsertUserProfile(ctx context.Context, profile *model.UserProfile) error {
	if profile.ClassIDs == nil {
		profile.ClassIDs = []string{}
	}
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO user_profiles (user_sub, iban, address, phone, class_ids, profile_complete, profile_skipped, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		 ON CONFLICT (user_sub) DO UPDATE
		   SET iban = EXCLUDED.iban,
		       address = EXCLUDED.address,
		       phone = EXCLUDED.phone,
		       class_ids = EXCLUDED.class_ids,
		       profile_complete = EXCLUDED.profile_complete,
		       profile_skipped = EXCLUDED.profile_skipped,
		       updated_at = NOW()`,
		profile.UserSub, profile.IBAN, profile.Address, profile.Phone, pq.Array(profile.ClassIDs),
		profile.ProfileComplete, profile.ProfileSkipped,
	)
	return err
}

// — school class CRUD —

func (p *Postgres) CreateSchoolClass(ctx context.Context, sc *model.SchoolClass) error {
	sc.ID = newID()
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	classroomID := ""
	if sc.Classroom != nil {
		classroomID = sc.Classroom.ID
	}
	schoolYearID := ""
	if sc.SchoolYear != nil {
		schoolYearID = sc.SchoolYear.ID
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO school_classes (id, name, shortcut, classroom_id, school_year_id, version)
		 VALUES ($1, $2, $3, $4, $5, 0)`,
		sc.ID, nullableString(sc.Name), nullableString(sc.Shortcut),
		nullableString(classroomID), nullableString(schoolYearID),
	); err != nil {
		return fmt.Errorf("inserting school class: %w", err)
	}
	if err := syncSchoolClassTeachers(ctx, tx, sc.ID, sc.Teachers); err != nil {
		return err
	}
	if err := syncSchoolClassStudents(ctx, tx, sc.ID, sc.Students); err != nil {
		return err
	}
	return tx.Commit()
}

func (p *Postgres) GetSchoolClass(ctx context.Context, id string) (*model.SchoolClass, error) {
	sc := &model.SchoolClass{}
	var shortcut, classroomID, schoolYearID sql.NullString
	err := p.db.QueryRowContext(ctx,
		`SELECT id, COALESCE(name,''), shortcut, classroom_id, school_year_id, version
		 FROM school_classes WHERE id = $1`, id,
	).Scan(&sc.ID, &sc.Name, &shortcut, &classroomID, &schoolYearID, &sc.Version)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying school class: %w", err)
	}
	sc.Shortcut = shortcut.String
	if classroomID.Valid {
		sc.Classroom = &model.Room{ID: classroomID.String}
	}
	if schoolYearID.Valid {
		sc.SchoolYear = &model.SchoolYear{ID: schoolYearID.String}
	}
	teachers, students, err := p.loadSchoolClassMembers(ctx, id)
	if err != nil {
		return nil, err
	}
	sc.Teachers = teachers
	sc.Students = students
	return sc, nil
}

func (p *Postgres) ListSchoolClasses(ctx context.Context) ([]*model.SchoolClass, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT id, COALESCE(name,''), shortcut, classroom_id, school_year_id, version
		 FROM school_classes ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("listing school classes: %w", err)
	}
	defer rows.Close()

	var classes []*model.SchoolClass
	for rows.Next() {
		sc := &model.SchoolClass{}
		var shortcut, classroomID, schoolYearID sql.NullString
		if err := rows.Scan(&sc.ID, &sc.Name, &shortcut, &classroomID, &schoolYearID, &sc.Version); err != nil {
			return nil, err
		}
		sc.Shortcut = shortcut.String
		if classroomID.Valid {
			sc.Classroom = &model.Room{ID: classroomID.String}
		}
		if schoolYearID.Valid {
			sc.SchoolYear = &model.SchoolYear{ID: schoolYearID.String}
		}
		sc.Teachers = []model.Teacher{}
		sc.Students = []model.Student{}
		classes = append(classes, sc)
	}
	return classes, rows.Err()
}

func (p *Postgres) UpdateSchoolClass(ctx context.Context, sc *model.SchoolClass) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	classroomID := ""
	if sc.Classroom != nil {
		classroomID = sc.Classroom.ID
	}
	schoolYearID := ""
	if sc.SchoolYear != nil {
		schoolYearID = sc.SchoolYear.ID
	}
	res, err := tx.ExecContext(ctx,
		`UPDATE school_classes SET name=$1, shortcut=$2, classroom_id=$3, school_year_id=$4, version=version+1
		 WHERE id=$5 AND version=$6`,
		nullableString(sc.Name), nullableString(sc.Shortcut),
		nullableString(classroomID), nullableString(schoolYearID),
		sc.ID, sc.Version,
	)
	if err != nil {
		return fmt.Errorf("updating school class: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		var exists bool
		_ = p.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM school_classes WHERE id=$1)`, sc.ID).Scan(&exists)
		if !exists {
			return fmt.Errorf("school class not found")
		}
		return ErrOptimisticLock
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM school_class_teachers WHERE school_class_id=$1`, sc.ID); err != nil {
		return fmt.Errorf("clearing school class teachers: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM school_class_students WHERE school_class_id=$1`, sc.ID); err != nil {
		return fmt.Errorf("clearing school class students: %w", err)
	}
	if err := syncSchoolClassTeachers(ctx, tx, sc.ID, sc.Teachers); err != nil {
		return err
	}
	if err := syncSchoolClassStudents(ctx, tx, sc.ID, sc.Students); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	sc.Version++
	return nil
}

func (p *Postgres) DeleteSchoolClass(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "school_classes", id)
}

func syncSchoolClassTeachers(ctx context.Context, tx *sql.Tx, classID string, teachers []model.Teacher) error {
	for _, t := range teachers {
		// If the teacher has a sub field (Keycloak subject ID), we need to look up the person_id
		var teacherID string
		if t.Sub != nil && *t.Sub != "" {
			// Look up person_id from persons table using sub
			err := tx.QueryRowContext(ctx,
				`SELECT id FROM persons WHERE sub = $1`,
				t.Sub,
			).Scan(&teacherID)
			if err != nil {
				if err == sql.ErrNoRows {
					subVal := ""
				if t.Sub != nil {
					subVal = *t.Sub
				}
				return fmt.Errorf("teacher with sub %s not found in local database. The teacher must be created first before assigning to school classes", subVal)
				}
				return fmt.Errorf("looking up teacher by sub: %w", err)
			}
		} else if t.ID != "" {
			// Fallback to using ID directly (for existing teachers)
			teacherID = t.ID
		} else {
			// Neither Sub nor ID is available - skip this teacher
			continue
		}

		// Validate that we have a teacher ID to work with
		if teacherID == "" {
			return fmt.Errorf("teacher has neither valid sub nor id field")
		}

		_, err := tx.ExecContext(ctx,
			`INSERT INTO school_class_teachers (school_class_id, teacher_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			classID, teacherID,
		)
		if err != nil {
			return fmt.Errorf("linking teacher to school class (class_id=%s, teacher_id=%s): %w", classID, teacherID, err)
		}
	}
	return nil
}

func syncSchoolClassStudents(ctx context.Context, tx *sql.Tx, classID string, students []model.Student) error {
	for _, s := range students {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO school_class_students (school_class_id, student_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			classID, s.ID,
		); err != nil {
			return fmt.Errorf("linking student to school class: %w", err)
		}
	}
	return nil
}

func (p *Postgres) loadSchoolClassMembers(ctx context.Context, classID string) (teachers []model.Teacher, students []model.Student, err error) {
	// Load teachers (basic info only)
	tRows, err := p.db.QueryContext(ctx, `
		SELECT p.id, p.name, p.prename, p.version
		FROM persons p
		JOIN teachers t ON t.person_id = p.id
		JOIN school_class_teachers sct ON sct.teacher_id = p.id
		WHERE sct.school_class_id = $1
		ORDER BY p.name`, classID,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("loading school class teachers: %w", err)
	}
	defer tRows.Close()
	for tRows.Next() {
		t := model.Teacher{}
		if err := tRows.Scan(&t.ID, &t.Name, &t.Prename, &t.Version); err != nil {
			return nil, nil, err
		}
		teachers = append(teachers, t)
	}
	if teachers == nil {
		teachers = []model.Teacher{}
	}

	// Load students (basic info only)
	sRows, err := p.db.QueryContext(ctx, `
		SELECT p.id, p.name, p.prename, p.version
		FROM persons p
		JOIN students s ON s.person_id = p.id
		JOIN school_class_students scs ON scs.student_id = p.id
		WHERE scs.school_class_id = $1
		ORDER BY p.name`, classID,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("loading school class students: %w", err)
	}
	defer sRows.Close()
	for sRows.Next() {
		s := model.Student{}
		if err := sRows.Scan(&s.ID, &s.Name, &s.Prename, &s.Version); err != nil {
			return nil, nil, err
		}
		s.Guardians = []model.Guardian{}
		s.PastSchoolClasses = []model.SchoolClass{}
		s.Grades = []model.Grade{}
		students = append(students, s)
	}
	if students == nil {
		students = []model.Student{}
	}
	return teachers, students, nil
}

// syncClassPersons and loadClassPersons are kept for backward compatibility
// with any existing data in class_persons table (not used by new handlers).
func syncClassPersons(ctx context.Context, tx *sql.Tx, classID string, persons []model.Person, role string) error {
	for i := range persons {
		p := &persons[i]
		if p.ID == "" {
			if p.Sub != nil && *p.Sub != "" {
				p.ID = *p.Sub
			} else {
				p.ID = newID()
			}
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO persons (id, name, sub) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, sub = EXCLUDED.sub`,
			p.ID, p.Name, p.Sub,
		); err != nil {
			return fmt.Errorf("upserting person: %w", err)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO class_persons (class_id, person_id, role) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
			classID, p.ID, role,
		); err != nil {
			return fmt.Errorf("inserting class person: %w", err)
		}
	}
	return nil
}

// nullableString returns nil for empty strings so they map to SQL NULL.
func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// nullableStringPtr returns nil for nil or empty string pointers so they map to SQL NULL.
func nullableStringPtr(s *string) interface{} {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}

// nullableFloat64 returns nil for nil pointers so they map to SQL NULL.
func nullableFloat64(f *float64) interface{} {
	if f == nil {
		return nil
	}
	return *f
}

// nullableInt returns nil for nil pointers so they map to SQL NULL.
func nullableInt(i *int) interface{} {
	if i == nil {
		return nil
	}
	return *i
}

// nullableJSON returns nil for empty byte slices so they map to SQL NULL.
func nullableJSON(b []byte) interface{} {
	if len(b) == 0 {
		return nil
	}
	return b
}

func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// optimisticCheck checks the result of an UPDATE WHERE id=$x AND version=$v.
// Returns ErrOptimisticLock if the row exists but version didn't match,
// or a "not found" error if the row doesn't exist at all.
// On success it increments *version to reflect the new DB value.
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

// deleteByID deletes a row by TEXT id and returns an error if not found.
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

// — user registry —

func (p *Postgres) CreateUserRegistry(ctx context.Context, user *model.UserRegistry) error {
	user.ID = newID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := p.db.ExecContext(ctx,
		`INSERT INTO user_registry 
		 (id, user_sub, person_id, email, registration_status, keycloak_roles, local_roles, 
		  last_role_sync, created_at, updated_at, last_login_at, created_by, updated_by, 
		  legacy_profile_id, legacy_request_id) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		user.ID, user.UserSub, nullableString(user.PersonID), nullableString(user.Email),
		user.RegistrationStatus, pq.Array(user.KeycloakRoles), pq.Array(user.LocalRoles),
		nullableTime(user.LastRoleSync), user.CreatedAt, user.UpdatedAt, nullableTime(user.LastLoginAt),
		nullableString(user.CreatedBy), nullableString(user.UpdatedBy),
		nullableString(user.LegacyProfileID), nullableString(user.LegacyRequestID),
	)
	return err
}

func (p *Postgres) GetUserRegistry(ctx context.Context, id string) (*model.UserRegistry, error) {
	user := &model.UserRegistry{}
	var personID, email, createdBy, updatedBy, legacyProfileID, legacyRequestID sql.NullString
	var lastRoleSync, lastLoginAt sql.NullTime

	err := p.db.QueryRowContext(ctx,
		`SELECT id, user_sub, person_id, email, registration_status, keycloak_roles, local_roles,
		 last_role_sync, created_at, updated_at, last_login_at, created_by, updated_by,
		 legacy_profile_id, legacy_request_id
		 FROM user_registry WHERE id = $1`, id,
	).Scan(&user.ID, &user.UserSub, &personID, &email, &user.RegistrationStatus,
		pq.Array(&user.KeycloakRoles), pq.Array(&user.LocalRoles),
		&lastRoleSync, &user.CreatedAt, &user.UpdatedAt, &lastLoginAt,
		&createdBy, &updatedBy, &legacyProfileID, &legacyRequestID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting user registry: %w", err)
	}

	user.PersonID = personID.String
	user.Email = email.String
	if lastRoleSync.Valid {
		t := time.Time(lastRoleSync.Time)
		user.LastRoleSync = &t
	}
	if lastLoginAt.Valid {
		t := time.Time(lastLoginAt.Time)
		user.LastLoginAt = &t
	}
	user.CreatedBy = createdBy.String
	user.UpdatedBy = updatedBy.String
	user.LegacyProfileID = legacyProfileID.String
	user.LegacyRequestID = legacyRequestID.String

	return user, nil
}

func (p *Postgres) GetUserRegistryBySub(ctx context.Context, userSub string) (*model.UserRegistry, error) {
	user := &model.UserRegistry{}
	var personID, email, createdBy, updatedBy, legacyProfileID, legacyRequestID sql.NullString
	var lastRoleSync, lastLoginAt sql.NullTime

	err := p.db.QueryRowContext(ctx,
		`SELECT id, user_sub, person_id, email, registration_status, keycloak_roles, local_roles,
		 last_role_sync, created_at, updated_at, last_login_at, created_by, updated_by,
		 legacy_profile_id, legacy_request_id
		 FROM user_registry WHERE user_sub = $1`, userSub,
	).Scan(&user.ID, &user.UserSub, &personID, &email, &user.RegistrationStatus,
		pq.Array(&user.KeycloakRoles), pq.Array(&user.LocalRoles),
		&lastRoleSync, &user.CreatedAt, &user.UpdatedAt, &lastLoginAt,
		&createdBy, &updatedBy, &legacyProfileID, &legacyRequestID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting user registry by sub: %w", err)
	}

	user.PersonID = personID.String
	user.Email = email.String
	if lastRoleSync.Valid {
		t := time.Time(lastRoleSync.Time)
		user.LastRoleSync = &t
	}
	if lastLoginAt.Valid {
		t := time.Time(lastLoginAt.Time)
		user.LastLoginAt = &t
	}
	user.CreatedBy = createdBy.String
	user.UpdatedBy = updatedBy.String
	user.LegacyProfileID = legacyProfileID.String
	user.LegacyRequestID = legacyRequestID.String

	return user, nil
}

func (p *Postgres) GetUserRegistryByPersonID(ctx context.Context, personID string) (*model.UserRegistry, error) {
	user := &model.UserRegistry{}
	var userSub, email, createdBy, updatedBy, legacyProfileID, legacyRequestID sql.NullString
	var lastRoleSync, lastLoginAt sql.NullTime

	err := p.db.QueryRowContext(ctx,
		`SELECT id, user_sub, email, registration_status, keycloak_roles, local_roles,
		 last_role_sync, created_at, updated_at, last_login_at, created_by, updated_by,
		 legacy_profile_id, legacy_request_id
		 FROM user_registry WHERE person_id = $1`, personID,
	).Scan(&user.ID, &userSub, &email, &user.RegistrationStatus,
		pq.Array(&user.KeycloakRoles), pq.Array(&user.LocalRoles),
		&lastRoleSync, &user.CreatedAt, &user.UpdatedAt, &lastLoginAt,
		&createdBy, &updatedBy, &legacyProfileID, &legacyRequestID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting user registry by person ID: %w", err)
	}

	user.UserSub = userSub.String
	user.Email = email.String
	if lastRoleSync.Valid {
		t := time.Time(lastRoleSync.Time)
		user.LastRoleSync = &t
	}
	if lastLoginAt.Valid {
		t := time.Time(lastLoginAt.Time)
		user.LastLoginAt = &t
	}
	user.CreatedBy = createdBy.String
	user.UpdatedBy = updatedBy.String
	user.LegacyProfileID = legacyProfileID.String
	user.LegacyRequestID = legacyRequestID.String

	return user, nil
}

func (p *Postgres) UpdateUserRegistry(ctx context.Context, user *model.UserRegistry) error {
	user.UpdatedAt = time.Now()

	_, err := p.db.ExecContext(ctx,
		`UPDATE user_registry SET
		 user_sub = $1, person_id = $2, email = $3, registration_status = $4,
		 keycloak_roles = $5, local_roles = $6, last_role_sync = $7,
		 updated_at = $8, last_login_at = $9, updated_by = $10,
		 legacy_profile_id = $11, legacy_request_id = $12
		 WHERE id = $13`,
		user.UserSub, nullableString(user.PersonID), nullableString(user.Email),
		user.RegistrationStatus, pq.Array(user.KeycloakRoles), pq.Array(user.LocalRoles),
		nullableTime(user.LastRoleSync), user.UpdatedAt, nullableTime(user.LastLoginAt),
		nullableString(user.UpdatedBy), nullableString(user.LegacyProfileID),
		nullableString(user.LegacyRequestID), user.ID,
	)
	return err
}

func (p *Postgres) ListUserRegistries(ctx context.Context) ([]*model.UserRegistry, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT id, user_sub, person_id, email, registration_status, keycloak_roles, local_roles,
		 last_role_sync, created_at, updated_at, last_login_at, created_by, updated_by,
		 legacy_profile_id, legacy_request_id
		 FROM user_registry ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing user registries: %w", err)
	}
	defer rows.Close()

	var users []*model.UserRegistry
	for rows.Next() {
		user := &model.UserRegistry{}
		var personID, email, createdBy, updatedBy, legacyProfileID, legacyRequestID sql.NullString
		var lastRoleSync, lastLoginAt sql.NullTime

		err := rows.Scan(&user.ID, &user.UserSub, &personID, &email, &user.RegistrationStatus,
			pq.Array(&user.KeycloakRoles), pq.Array(&user.LocalRoles),
			&lastRoleSync, &user.CreatedAt, &user.UpdatedAt, &lastLoginAt,
			&createdBy, &updatedBy, &legacyProfileID, &legacyRequestID,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning user registry: %w", err)
		}

		user.PersonID = personID.String
		user.Email = email.String
		if lastRoleSync.Valid {
			t := time.Time(lastRoleSync.Time)
			user.LastRoleSync = &t
		}
		if lastLoginAt.Valid {
			t := time.Time(lastLoginAt.Time)
			user.LastLoginAt = &t
		}
		user.CreatedBy = createdBy.String
		user.UpdatedBy = updatedBy.String
		user.LegacyProfileID = legacyProfileID.String
		user.LegacyRequestID = legacyRequestID.String

		users = append(users, user)
	}

	return users, nil
}

// — registration workflow —

func (p *Postgres) CreateRegistrationWorkflow(ctx context.Context, workflow *model.RegistrationWorkflow) error {
	workflow.ID = newID()
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = time.Now()

	// Convert request data to JSONB
	requestDataJSON, err := json.Marshal(workflow.RequestData)
	if err != nil {
		return fmt.Errorf("marshaling request data: %w", err)
	}

	// Convert auto approval rules to JSONB
	var autoApprovalRulesJSON []byte
	if workflow.AutoApprovalRules != nil {
		autoApprovalRulesJSON, err = json.Marshal(workflow.AutoApprovalRules)
		if err != nil {
			return fmt.Errorf("marshaling auto approval rules: %w", err)
		}
	}

	_, err = p.db.ExecContext(ctx,
		`INSERT INTO registration_workflow
		 (id, user_id, request_type, request_data, desired_roles, current_roles,
		  approval_status, requires_manual_approval, auto_approval_rules,
		  pending_teacher_id, pending_student_id, pending_guardian_id,
		  created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		workflow.ID, workflow.UserID, workflow.RequestType, requestDataJSON,
		pq.Array(workflow.DesiredRoles), pq.Array(workflow.CurrentRoles),
		workflow.ApprovalStatus, workflow.RequiresManualApproval, nullableJSON(autoApprovalRulesJSON),
		nullableString(workflow.PendingTeacherID), nullableString(workflow.PendingStudentID),
		nullableString(workflow.PendingGuardianID), workflow.CreatedAt, workflow.UpdatedAt,
	)
	return err
}

func (p *Postgres) GetRegistrationWorkflow(ctx context.Context, id string) (*model.RegistrationWorkflow, error) {
	workflow := &model.RegistrationWorkflow{}
	var userID, requestType, approvalStatus sql.NullString
	var approvalBy, rejectionBy, rejectionReason, pendingTeacherID, pendingStudentID, pendingGuardianID sql.NullString
	var approvalNotes sql.NullString // Added for NULL handling
	var approvalAt, rejectionAt, completedAt sql.NullTime
	var requestDataJSON, autoApprovalRulesJSON []byte

	err := p.db.QueryRowContext(ctx,
		`SELECT id, user_id, request_type, request_data, desired_roles, current_roles,
		 approval_status, approval_by, approval_at, approval_notes, rejection_reason,
		 rejection_by, rejection_at, requires_manual_approval, auto_approval_rules,
		 pending_teacher_id, pending_student_id, pending_guardian_id,
		 created_at, updated_at, completed_at
		 FROM registration_workflow WHERE id = $1`, id,
	).Scan(&workflow.ID, &userID, &requestType, &requestDataJSON, pq.Array(&workflow.DesiredRoles),
		pq.Array(&workflow.CurrentRoles), &approvalStatus, &approvalBy, &approvalAt,
		&approvalNotes, &rejectionReason, &rejectionBy, &rejectionAt,
		&workflow.RequiresManualApproval, &autoApprovalRulesJSON, &pendingTeacherID,
		&pendingStudentID, &pendingGuardianID, &workflow.CreatedAt, &workflow.UpdatedAt, &completedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting registration workflow: %w", err)
	}

	workflow.UserID = userID.String
	workflow.ApprovalNotes = approvalNotes.String // Added NULL-safe conversion
	workflow.RequestType = requestType.String
	workflow.ApprovalStatus = approvalStatus.String
	workflow.ApprovalBy = approvalBy.String
	workflow.RejectionReason = rejectionReason.String
	workflow.RejectionBy = rejectionBy.String
	workflow.PendingTeacherID = pendingTeacherID.String
	workflow.PendingStudentID = pendingStudentID.String
	workflow.PendingGuardianID = pendingGuardianID.String

	if approvalAt.Valid {
		t := time.Time(approvalAt.Time)
		workflow.ApprovalAt = &t
	}
	if rejectionAt.Valid {
		t := time.Time(rejectionAt.Time)
		workflow.RejectionAt = &t
	}
	if completedAt.Valid {
		t := time.Time(completedAt.Time)
		workflow.CompletedAt = &t
	}

	// Unmarshal JSON fields
	if len(requestDataJSON) > 0 {
		if err := json.Unmarshal(requestDataJSON, &workflow.RequestData); err != nil {
			return nil, fmt.Errorf("unmarshaling request data: %w", err)
		}
	}

	if len(autoApprovalRulesJSON) > 0 {
		if err := json.Unmarshal(autoApprovalRulesJSON, &workflow.AutoApprovalRules); err != nil {
			return nil, fmt.Errorf("unmarshaling auto approval rules: %w", err)
		}
	}

	return workflow, nil
}

func (p *Postgres) GetLatestWorkflowByUserID(ctx context.Context, userID string) (*model.RegistrationWorkflow, error) {
	workflow := &model.RegistrationWorkflow{}
	var wUserID, requestType, approvalStatus sql.NullString
	var approvalBy, rejectionBy, rejectionReason, pendingTeacherID, pendingStudentID, pendingGuardianID sql.NullString
	var approvalNotes sql.NullString
	var approvalAt, rejectionAt, completedAt sql.NullTime
	var requestDataJSON, autoApprovalRulesJSON []byte

	err := p.db.QueryRowContext(ctx,
		`SELECT id, user_id, request_type, request_data, desired_roles, current_roles,
		 approval_status, approval_by, approval_at, approval_notes, rejection_reason,
		 rejection_by, rejection_at, requires_manual_approval, auto_approval_rules,
		 pending_teacher_id, pending_student_id, pending_guardian_id,
		 created_at, updated_at, completed_at
		 FROM registration_workflow WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`, userID,
	).Scan(&workflow.ID, &wUserID, &requestType, &requestDataJSON, pq.Array(&workflow.DesiredRoles),
		pq.Array(&workflow.CurrentRoles), &approvalStatus, &approvalBy, &approvalAt,
		&approvalNotes, &rejectionReason, &rejectionBy, &rejectionAt,
		&workflow.RequiresManualApproval, &autoApprovalRulesJSON, &pendingTeacherID,
		&pendingStudentID, &pendingGuardianID, &workflow.CreatedAt, &workflow.UpdatedAt, &completedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting latest workflow by user ID: %w", err)
	}

	workflow.UserID = wUserID.String
	workflow.RequestType = requestType.String
	workflow.ApprovalStatus = approvalStatus.String
	workflow.ApprovalBy = approvalBy.String
	workflow.ApprovalNotes = approvalNotes.String
	workflow.RejectionReason = rejectionReason.String
	workflow.RejectionBy = rejectionBy.String
	workflow.PendingTeacherID = pendingTeacherID.String
	workflow.PendingStudentID = pendingStudentID.String
	workflow.PendingGuardianID = pendingGuardianID.String

	if approvalAt.Valid {
		t := time.Time(approvalAt.Time)
		workflow.ApprovalAt = &t
	}
	if rejectionAt.Valid {
		t := time.Time(rejectionAt.Time)
		workflow.RejectionAt = &t
	}
	if completedAt.Valid {
		t := time.Time(completedAt.Time)
		workflow.CompletedAt = &t
	}

	if len(requestDataJSON) > 0 {
		if err := json.Unmarshal(requestDataJSON, &workflow.RequestData); err != nil {
			return nil, fmt.Errorf("unmarshaling request data: %w", err)
		}
	}
	if len(autoApprovalRulesJSON) > 0 {
		if err := json.Unmarshal(autoApprovalRulesJSON, &workflow.AutoApprovalRules); err != nil {
			return nil, fmt.Errorf("unmarshaling auto approval rules: %w", err)
		}
	}

	return workflow, nil
}

func (p *Postgres) ListRegistrationWorkflows(ctx context.Context, statusFilter string) ([]*model.RegistrationWorkflow, error) {
	var query string
	var params []interface{}

	if statusFilter != "" {
		query = `SELECT id, user_id, request_type, request_data, desired_roles, current_roles,
		 approval_status, approval_by, approval_at, approval_notes, rejection_reason,
		 rejection_by, rejection_at, requires_manual_approval, auto_approval_rules,
		 pending_teacher_id, pending_student_id, pending_guardian_id,
		 created_at, updated_at, completed_at
		 FROM registration_workflow WHERE approval_status = $1 ORDER BY created_at DESC`
		params = append(params, statusFilter)
	} else {
		query = `SELECT id, user_id, request_type, request_data, desired_roles, current_roles,
		 approval_status, approval_by, approval_at, approval_notes, rejection_reason,
		 rejection_by, rejection_at, requires_manual_approval, auto_approval_rules,
		 pending_teacher_id, pending_student_id, pending_guardian_id,
		 created_at, updated_at, completed_at
		 FROM registration_workflow ORDER BY created_at DESC`
	}

	rows, err := p.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("listing registration workflows: %w", err)
	}
	defer rows.Close()

	var workflows []*model.RegistrationWorkflow
	for rows.Next() {
		workflow := &model.RegistrationWorkflow{}
		var userID, requestType, approvalStatus sql.NullString
		var approvalBy, rejectionBy, rejectionReason, pendingTeacherID, pendingStudentID, pendingGuardianID sql.NullString
		var approvalNotes sql.NullString // Added for NULL handling
		var approvalAt, rejectionAt, completedAt sql.NullTime
		var requestDataJSON, autoApprovalRulesJSON []byte

		err := rows.Scan(&workflow.ID, &userID, &requestType, &requestDataJSON, pq.Array(&workflow.DesiredRoles),
			pq.Array(&workflow.CurrentRoles), &approvalStatus, &approvalBy, &approvalAt,
			&approvalNotes, &rejectionReason, &rejectionBy, &rejectionAt,
			&workflow.RequiresManualApproval, &autoApprovalRulesJSON, &pendingTeacherID,
			&pendingStudentID, &pendingGuardianID, &workflow.CreatedAt, &workflow.UpdatedAt, &completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning registration workflow: %w", err)
		}

		workflow.UserID = userID.String
		workflow.RequestType = requestType.String
		workflow.ApprovalStatus = approvalStatus.String
		workflow.ApprovalBy = approvalBy.String
		workflow.ApprovalNotes = approvalNotes.String // Added NULL-safe conversion
		workflow.RejectionReason = rejectionReason.String
		workflow.RejectionBy = rejectionBy.String
		workflow.PendingTeacherID = pendingTeacherID.String
		workflow.PendingStudentID = pendingStudentID.String
		workflow.PendingGuardianID = pendingGuardianID.String

		if approvalAt.Valid {
			t := time.Time(approvalAt.Time)
			workflow.ApprovalAt = &t
		}
		if rejectionAt.Valid {
			t := time.Time(rejectionAt.Time)
			workflow.RejectionAt = &t
		}
		if completedAt.Valid {
			t := time.Time(completedAt.Time)
			workflow.CompletedAt = &t
		}

		// Unmarshal JSON fields
		if len(requestDataJSON) > 0 {
			if err := json.Unmarshal(requestDataJSON, &workflow.RequestData); err != nil {
				return nil, fmt.Errorf("unmarshaling request data: %w", err)
			}
		}

		if len(autoApprovalRulesJSON) > 0 {
			if err := json.Unmarshal(autoApprovalRulesJSON, &workflow.AutoApprovalRules); err != nil {
				return nil, fmt.Errorf("unmarshaling auto approval rules: %w", err)
			}
		}

		workflows = append(workflows, workflow)
	}

	return workflows, nil
}

func (p *Postgres) UpdateRegistrationWorkflow(ctx context.Context, workflow *model.RegistrationWorkflow) error {
	workflow.UpdatedAt = time.Now()

	// Convert request data to JSONB
	requestDataJSON, err := json.Marshal(workflow.RequestData)
	if err != nil {
		return fmt.Errorf("marshaling request data: %w", err)
	}

	// Convert auto approval rules to JSONB; pass nil so Postgres stores NULL when empty.
	var autoApprovalRulesJSON interface{}
	if workflow.AutoApprovalRules != nil {
		b, merr := json.Marshal(workflow.AutoApprovalRules)
		if merr != nil {
			return fmt.Errorf("marshaling auto approval rules: %w", merr)
		}
		autoApprovalRulesJSON = b
	}

	_, err = p.db.ExecContext(ctx,
		`UPDATE registration_workflow SET
		 user_id = $1, request_type = $2, request_data = $3, desired_roles = $4,
		 current_roles = $5, approval_status = $6, approval_by = $7, approval_at = $8,
		 approval_notes = $9, rejection_reason = $10, rejection_by = $11, rejection_at = $12,
		 requires_manual_approval = $13, auto_approval_rules = $14,
		 pending_teacher_id = $15, pending_student_id = $16, pending_guardian_id = $17,
		 updated_at = $18, completed_at = $19
		 WHERE id = $20`,
		nullableString(workflow.UserID), workflow.RequestType, requestDataJSON,
		pq.Array(workflow.DesiredRoles), pq.Array(workflow.CurrentRoles),
		workflow.ApprovalStatus, nullableString(workflow.ApprovalBy), nullableTime(workflow.ApprovalAt),
		nullableString(workflow.ApprovalNotes), nullableString(workflow.RejectionReason),
		nullableString(workflow.RejectionBy), nullableTime(workflow.RejectionAt),
		workflow.RequiresManualApproval, autoApprovalRulesJSON,
		nullableString(workflow.PendingTeacherID), nullableString(workflow.PendingStudentID),
		nullableString(workflow.PendingGuardianID), workflow.UpdatedAt, nullableTime(workflow.CompletedAt),
		workflow.ID,
	)
	return err
}
