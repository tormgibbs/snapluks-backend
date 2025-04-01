package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/tormgibbs/snapluks-backend/internal/validator"
)

type Provider struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Latitude  *float64  `json:"latitude"`
	Longitude *float64  `json:"longitude"`
	CreatedAt time.Time `json:"-"`
	Version   int32     `json:"version"`
}

type ProviderModel struct {
	DB *sql.DB
}

func (pm ProviderModel) Insert(p *Provider) error {
	query := `
		INSERT INTO providers (name, address, latitude, longitude)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version
	`

	args := []any{p.Name, p.Address, p.Latitude, p.Longitude}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return pm.DB.QueryRowContext(ctx, query, args...).Scan(&p.ID, &p.CreatedAt, &p.Version)
}

func (pm ProviderModel) GetAll(name, address string, latitude, longitude float64, filters Filters) ([]*Provider, error) {
	query := `
		SELECT id, created_at, name, address, latitude, longitude, version
		FROM providers
		ORDER BY id;
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := pm.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*Provider

	for rows.Next() {
		var provider Provider
		err := rows.Scan(
			&provider.ID,
			&provider.CreatedAt,
			&provider.Name,
			&provider.Address,
			&provider.Latitude,
			&provider.Longitude,
			&provider.Version,
		)
		if err != nil {
			return nil, err
		}
		providers = append(providers, &provider)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return providers, nil
}

func (pm ProviderModel) Get(id int64) (*Provider, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, name, address, latitude, longitude, created_at, version
		FROM providers
		WHERE id = $1
	`

	var provider Provider

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := pm.DB.QueryRowContext(ctx, query, id).Scan(
		&provider.ID,
		&provider.Name,
		&provider.Address,
		&provider.Latitude,
		&provider.Longitude,
		&provider.CreatedAt,
		&provider.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &provider, nil
}

func (pm ProviderModel) Update(p *Provider) error {
	query := `
		UPDATE providers
		SET name = $1, address = $2, latitude = $3, longitude = $4
		WHERE id = $5 and version = $6
		RETURNING version
	`
	args := []any{
		p.Name,
		p.Address,
		p.Latitude,
		p.Longitude,
		p.ID,
		p.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := pm.DB.QueryRowContext(ctx, query, args...).Scan(&p.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (pm ProviderModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM providers
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := pm.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func ValidateProvider(v *validator.Validator, p *Provider) {
	v.Check(p.Name != "", "name", "must be provided")
	v.Check(p.Address != "", "address", "must be provided")

	// Latitude validation
	v.Check(p.Latitude != nil, "latitude", "must be provided")
	if p.Latitude != nil {
		v.Check(
			*p.Latitude >= -90 && *p.Latitude <= 90, "latitude", "must be between -90 and 90",
		)
	}

	// Longitude validation
	v.Check(p.Longitude != nil, "longitude", "must be provided")
	if p.Longitude != nil {
		v.Check(
			*p.Longitude >= -180 && *p.Longitude <= 180, "longitude", "must be between -180 and 180",
		)
	}
}
