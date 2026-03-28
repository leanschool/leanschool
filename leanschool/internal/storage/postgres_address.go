package storage

import (
	"context"
	"database/sql"
	"fmt"

	model "github.com/Joel-Haeberli/leanschool-model"
)

// ── PostalCode ────────────────────────────────────────────────────────────────

func (p *Postgres) CreatePostalCode(ctx context.Context, pc *model.PostalCode) error {
	pc.ID = newID()
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO postal_codes (id, number, city, version) VALUES ($1, $2, $3, 0)`,
		pc.ID, pc.Number, pc.City,
	)
	return err
}

func (p *Postgres) GetPostalCode(ctx context.Context, id string) (*model.PostalCode, error) {
	pc := &model.PostalCode{}
	err := p.db.QueryRowContext(ctx,
		`SELECT id, number, city, version FROM postal_codes WHERE id = $1`, id,
	).Scan(&pc.ID, &pc.Number, &pc.City, &pc.Version)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying postal code: %w", err)
	}
	return pc, nil
}

func (p *Postgres) ListPostalCodes(ctx context.Context) ([]*model.PostalCode, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT id, number, city, version FROM postal_codes ORDER BY number`)
	if err != nil {
		return nil, fmt.Errorf("listing postal codes: %w", err)
	}
	defer rows.Close()
	var pcs []*model.PostalCode
	for rows.Next() {
		pc := &model.PostalCode{}
		if err := rows.Scan(&pc.ID, &pc.Number, &pc.City, &pc.Version); err != nil {
			return nil, err
		}
		pcs = append(pcs, pc)
	}
	return pcs, rows.Err()
}

func (p *Postgres) UpdatePostalCode(ctx context.Context, pc *model.PostalCode) error {
	res, err := p.db.ExecContext(ctx,
		`UPDATE postal_codes SET number=$1, city=$2, version=version+1 WHERE id=$3 AND version=$4`,
		pc.Number, pc.City, pc.ID, pc.Version,
	)
	if err != nil {
		return fmt.Errorf("updating postal code: %w", err)
	}
	return optimisticCheck(ctx, p.db, res, "postal_codes", pc.ID, &pc.Version)
}

func (p *Postgres) DeletePostalCode(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "postal_codes", id)
}

// ── City ──────────────────────────────────────────────────────────────────────

func (p *Postgres) CreateCity(ctx context.Context, c *model.City) error {
	c.ID = newID()
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO cities (id, name, version) VALUES ($1, $2, 0)`, c.ID, c.Name,
	); err != nil {
		return fmt.Errorf("inserting city: %w", err)
	}
	if err := syncCityPostalCodes(ctx, tx, c.ID, c.PostalCodes); err != nil {
		return err
	}
	return tx.Commit()
}

func (p *Postgres) GetCity(ctx context.Context, id string) (*model.City, error) {
	c := &model.City{}
	err := p.db.QueryRowContext(ctx,
		`SELECT id, name, version FROM cities WHERE id = $1`, id,
	).Scan(&c.ID, &c.Name, &c.Version)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying city: %w", err)
	}
	pcs, err := loadCityPostalCodes(ctx, p.db, id)
	if err != nil {
		return nil, err
	}
	c.PostalCodes = pcs
	return c, nil
}

func (p *Postgres) ListCities(ctx context.Context) ([]*model.City, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT id, name, version FROM cities ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("listing cities: %w", err)
	}
	defer rows.Close()
	var cities []*model.City
	for rows.Next() {
		c := &model.City{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Version); err != nil {
			return nil, err
		}
		cities = append(cities, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for _, c := range cities {
		pcs, err := loadCityPostalCodes(ctx, p.db, c.ID)
		if err != nil {
			return nil, err
		}
		c.PostalCodes = pcs
	}
	return cities, nil
}

func (p *Postgres) UpdateCity(ctx context.Context, c *model.City) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx,
		`UPDATE cities SET name=$1, version=version+1 WHERE id=$2 AND version=$3`,
		c.Name, c.ID, c.Version,
	)
	if err != nil {
		return fmt.Errorf("updating city: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		var exists bool
		_ = p.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM cities WHERE id=$1)`, c.ID).Scan(&exists)
		if !exists {
			return fmt.Errorf("city not found")
		}
		return ErrOptimisticLock
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM city_postal_codes WHERE city_id=$1`, c.ID); err != nil {
		return fmt.Errorf("clearing city postal codes: %w", err)
	}
	if err := syncCityPostalCodes(ctx, tx, c.ID, c.PostalCodes); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	c.Version++
	return nil
}

func (p *Postgres) DeleteCity(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "cities", id)
}

func syncCityPostalCodes(ctx context.Context, tx *sql.Tx, cityID string, pcs []model.PostalCode) error {
	for _, pc := range pcs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO city_postal_codes (city_id, postal_code_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			cityID, pc.ID,
		); err != nil {
			return fmt.Errorf("linking postal code %s to city: %w", pc.ID, err)
		}
	}
	return nil
}

func loadCityPostalCodes(ctx context.Context, db *sql.DB, cityID string) ([]model.PostalCode, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT pc.id, pc.number, pc.city, pc.version
		FROM postal_codes pc
		JOIN city_postal_codes cpc ON cpc.postal_code_id = pc.id
		WHERE cpc.city_id = $1
		ORDER BY pc.number`, cityID,
	)
	if err != nil {
		return nil, fmt.Errorf("loading city postal codes: %w", err)
	}
	defer rows.Close()
	var pcs []model.PostalCode
	for rows.Next() {
		var pc model.PostalCode
		if err := rows.Scan(&pc.ID, &pc.Number, &pc.City, &pc.Version); err != nil {
			return nil, err
		}
		pcs = append(pcs, pc)
	}
	if pcs == nil {
		pcs = []model.PostalCode{}
	}
	return pcs, rows.Err()
}

// ── Address ───────────────────────────────────────────────────────────────────

func (p *Postgres) CreateAddress(ctx context.Context, a *model.Address) error {
	a.ID = newID()
	pcID := ""
	if a.PostalCode != nil {
		pcID = a.PostalCode.ID
	}
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO addresses (id, street, number, postal_code_id, version) VALUES ($1, $2, $3, $4, 0)`,
		a.ID, nullableStringPtr(a.Street), nullableStringPtr(a.Number), nullableString(pcID),
	)
	return err
}

func (p *Postgres) GetAddress(ctx context.Context, id string) (*model.Address, error) {
	a := &model.Address{}
	var street, number sql.NullString
	var pcID sql.NullString
	var pcNo sql.NullInt64
	var pcCity sql.NullString
	var pcVer sql.NullInt64
	err := p.db.QueryRowContext(ctx, `
		SELECT a.id, a.street, a.number, a.version,
		       pc.id, pc.number, pc.city, pc.version
		FROM addresses a
		LEFT JOIN postal_codes pc ON pc.id = a.postal_code_id
		WHERE a.id = $1`, id,
	).Scan(
		&a.ID, &street, &number, &a.Version,
		&pcID, &pcNo, &pcCity, &pcVer,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying address: %w", err)
	}
	if street.Valid {
		a.Street = &street.String
	}
	if number.Valid {
		a.Number = &number.String
	}
	if pcID.Valid {
		n := int(pcNo.Int64)
		a.PostalCode = &model.PostalCode{ID: pcID.String, Number: n,City: pcCity.String, Version: int(pcVer.Int64)}
	}
	return a, nil
}

func (p *Postgres) ListAddresses(ctx context.Context) ([]*model.Address, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT a.id, a.street, a.number, a.version,
		       pc.id, pc.number, pc.city, pc.version
		FROM addresses a
		LEFT JOIN postal_codes pc ON pc.id = a.postal_code_id
		ORDER BY a.id`)
	if err != nil {
		return nil, fmt.Errorf("listing addresses: %w", err)
	}
	defer rows.Close()
	var addrs []*model.Address
	for rows.Next() {
		a := &model.Address{}
		var street, number sql.NullString
		var pcID sql.NullString
		var pcNo sql.NullInt64
		var pcCity sql.NullString
		var pcVer sql.NullInt64
		if err := rows.Scan(&a.ID, &street, &number, &a.Version, &pcID, &pcNo, &pcCity, &pcVer); err != nil {
			return nil, err
		}
		if street.Valid {
			a.Street = &street.String
		}
		if number.Valid {
			a.Number = &number.String
		}
		if pcID.Valid {
			n := int(pcNo.Int64)
			a.PostalCode = &model.PostalCode{ID: pcID.String, Number: n,City: pcCity.String, Version: int(pcVer.Int64)}
		}
		addrs = append(addrs, a)
	}
	return addrs, rows.Err()
}

func (p *Postgres) UpdateAddress(ctx context.Context, a *model.Address) error {
	pcID := ""
	if a.PostalCode != nil {
		pcID = a.PostalCode.ID
	}
	res, err := p.db.ExecContext(ctx,
		`UPDATE addresses SET street=$1, number=$2, postal_code_id=$3, version=version+1 WHERE id=$4 AND version=$5`,
		nullableStringPtr(a.Street), nullableStringPtr(a.Number), nullableString(pcID), a.ID, a.Version,
	)
	if err != nil {
		return fmt.Errorf("updating address: %w", err)
	}
	return optimisticCheck(ctx, p.db, res, "addresses", a.ID, &a.Version)
}

func (p *Postgres) DeleteAddress(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "addresses", id)
}
