package storage

import (
	"context"
	"database/sql"
	"fmt"

	model "github.com/Joel-Haeberli/leanschool-model"
)

// ── Location ─────────────────────────────────────────────────────────────────

func (p *Postgres) CreateLocation(ctx context.Context, l *model.Location) error {
	l.ID = newID()
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO locations (id, lon, lat, version) VALUES ($1, $2, $3, 0)`,
		l.ID, nullableFloat64(l.Lon), nullableFloat64(l.Lat),
	)
	return err
}

func (p *Postgres) GetLocation(ctx context.Context, id string) (*model.Location, error) {
	l := &model.Location{}
	var lon, lat sql.NullFloat64
	err := p.db.QueryRowContext(ctx,
		`SELECT id, lon, lat, version FROM locations WHERE id = $1`, id,
	).Scan(&l.ID, &lon, &lat, &l.Version)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying location: %w", err)
	}
	if lon.Valid {
		v := lon.Float64
		l.Lon = &v
	}
	if lat.Valid {
		v := lat.Float64
		l.Lat = &v
	}
	return l, nil
}

func (p *Postgres) ListLocations(ctx context.Context) ([]*model.Location, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT id, lon, lat, version FROM locations ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("listing locations: %w", err)
	}
	defer rows.Close()
	var locs []*model.Location
	for rows.Next() {
		l := &model.Location{}
		var lon, lat sql.NullFloat64
		if err := rows.Scan(&l.ID, &lon, &lat, &l.Version); err != nil {
			return nil, err
		}
		if lon.Valid {
			v := lon.Float64
			l.Lon = &v
		}
		if lat.Valid {
			v := lat.Float64
			l.Lat = &v
		}
		locs = append(locs, l)
	}
	return locs, rows.Err()
}

func (p *Postgres) UpdateLocation(ctx context.Context, l *model.Location) error {
	res, err := p.db.ExecContext(ctx,
		`UPDATE locations SET lon=$1, lat=$2, version=version+1 WHERE id=$3 AND version=$4`,
		nullableFloat64(l.Lon), nullableFloat64(l.Lat), l.ID, l.Version,
	)
	if err != nil {
		return fmt.Errorf("updating location: %w", err)
	}
	return optimisticCheck(ctx, p.db, res, "locations", l.ID, &l.Version)
}

func (p *Postgres) DeleteLocation(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "locations", id)
}

// ── Building ──────────────────────────────────────────────────────────────────

func (p *Postgres) CreateBuilding(ctx context.Context, b *model.Building) error {
	b.ID = newID()
	locID := ""
	if b.Location != nil {
		locID = b.Location.ID
	}
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO buildings (id, name, location_id, version) VALUES ($1, $2, $3, 0)`,
		b.ID, b.Name, nullableString(locID),
	)
	return err
}

func (p *Postgres) GetBuilding(ctx context.Context, id string) (*model.Building, error) {
	b := &model.Building{}
	var locID sql.NullString
	var locLon, locLat sql.NullFloat64
	var locVer sql.NullInt64
	err := p.db.QueryRowContext(ctx, `
		SELECT b.id, b.name, b.version,
		       l.id, l.lon, l.lat, l.version
		FROM buildings b
		LEFT JOIN locations l ON l.id = b.location_id
		WHERE b.id = $1`, id,
	).Scan(&b.ID, &b.Name, &b.Version, &locID, &locLon, &locLat, &locVer)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying building: %w", err)
	}
	if locID.Valid {
		lon := locLon.Float64
		lat := locLat.Float64
		b.Location = &model.Location{ID: locID.String, Lon: &lon, Lat: &lat, Version: int(locVer.Int64)}
	}
	return b, nil
}

func (p *Postgres) ListBuildings(ctx context.Context) ([]*model.Building, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT b.id, b.name, b.version,
		       l.id, l.lon, l.lat, l.version
		FROM buildings b
		LEFT JOIN locations l ON l.id = b.location_id
		ORDER BY b.name`)
	if err != nil {
		return nil, fmt.Errorf("listing buildings: %w", err)
	}
	defer rows.Close()
	var buildings []*model.Building
	for rows.Next() {
		b := &model.Building{}
		var locID sql.NullString
		var locLon, locLat sql.NullFloat64
		var locVer sql.NullInt64
		if err := rows.Scan(&b.ID, &b.Name, &b.Version, &locID, &locLon, &locLat, &locVer); err != nil {
			return nil, err
		}
		if locID.Valid {
			lon := locLon.Float64
			lat := locLat.Float64
			b.Location = &model.Location{ID: locID.String, Lon: &lon, Lat: &lat, Version: int(locVer.Int64)}
		}
		buildings = append(buildings, b)
	}
	return buildings, rows.Err()
}

func (p *Postgres) UpdateBuilding(ctx context.Context, b *model.Building) error {
	locID := ""
	if b.Location != nil {
		locID = b.Location.ID
	}
	res, err := p.db.ExecContext(ctx,
		`UPDATE buildings SET name=$1, location_id=$2, version=version+1 WHERE id=$3 AND version=$4`,
		b.Name, nullableString(locID), b.ID, b.Version,
	)
	if err != nil {
		return fmt.Errorf("updating building: %w", err)
	}
	return optimisticCheck(ctx, p.db, res, "buildings", b.ID, &b.Version)
}

func (p *Postgres) DeleteBuilding(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "buildings", id)
}

// ── Room ──────────────────────────────────────────────────────────────────────

func (p *Postgres) CreateRoom(ctx context.Context, r *model.Room) error {
	r.ID = newID()
	buildingID := ""
	if r.Building != nil {
		buildingID = r.Building.ID
	}
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO rooms (id, name, building_id, room_type, version) VALUES ($1, $2, $3, $4, 0)`,
		r.ID, r.Name, nullableString(buildingID), nullableString(roomTypeStr(r.RoomType)),
	)
	return err
}

func (p *Postgres) GetRoom(ctx context.Context, id string) (*model.Room, error) {
	r := &model.Room{}
	var bID sql.NullString
	var bName sql.NullString
	var bVer sql.NullInt64
	var roomType sql.NullString
	err := p.db.QueryRowContext(ctx, `
		SELECT r.id, r.name, r.room_type, r.version,
		       b.id, b.name, b.version
		FROM rooms r
		LEFT JOIN buildings b ON b.id = r.building_id
		WHERE r.id = $1`, id,
	).Scan(&r.ID, &r.Name, &roomType, &r.Version, &bID, &bName, &bVer)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying room: %w", err)
	}
	if roomType.Valid {
		rt := model.RoomType(roomType.String)
		r.RoomType = &rt
	}
	if bID.Valid {
		r.Building = &model.Building{ID: bID.String, Name: bName.String, Version: int(bVer.Int64)}
	}
	return r, nil
}

func (p *Postgres) ListRooms(ctx context.Context) ([]*model.Room, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT r.id, r.name, r.room_type, r.version,
		       b.id, b.name, b.version
		FROM rooms r
		LEFT JOIN buildings b ON b.id = r.building_id
		ORDER BY r.name`)
	if err != nil {
		return nil, fmt.Errorf("listing rooms: %w", err)
	}
	defer rows.Close()
	var rooms []*model.Room
	for rows.Next() {
		r := &model.Room{}
		var bID sql.NullString
		var bName sql.NullString
		var bVer sql.NullInt64
		var roomType sql.NullString
		if err := rows.Scan(&r.ID, &r.Name, &roomType, &r.Version, &bID, &bName, &bVer); err != nil {
			return nil, err
		}
		if roomType.Valid {
			rt := model.RoomType(roomType.String)
		r.RoomType = &rt
		}
		if bID.Valid {
			r.Building = &model.Building{ID: bID.String, Name: bName.String, Version: int(bVer.Int64)}
		}
		rooms = append(rooms, r)
	}
	return rooms, rows.Err()
}

func (p *Postgres) UpdateRoom(ctx context.Context, r *model.Room) error {
	buildingID := ""
	if r.Building != nil {
		buildingID = r.Building.ID
	}
	res, err := p.db.ExecContext(ctx,
		`UPDATE rooms SET name=$1, building_id=$2, room_type=$3, version=version+1 WHERE id=$4 AND version=$5`,
		r.Name, nullableString(buildingID), nullableString(roomTypeStr(r.RoomType)), r.ID, r.Version,
	)
	if err != nil {
		return fmt.Errorf("updating room: %w", err)
	}
	return optimisticCheck(ctx, p.db, res, "rooms", r.ID, &r.Version)
}

func (p *Postgres) DeleteRoom(ctx context.Context, id string) error {
	return deleteByID(ctx, p.db, "rooms", id)
}

// roomTypeStr converts a *model.RoomType to its string representation for SQL storage.
func roomTypeStr(rt *model.RoomType) string {
	if rt == nil {
		return ""
	}
	return string(*rt)
}
