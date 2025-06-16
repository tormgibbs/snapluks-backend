package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/tormgibbs/snapluks-backend/internal/validator"
)

type ProviderModel struct {
	DB *sql.DB
}

type Provider struct {
	ID          int     `json:"id"`
	UserID      int     `json:"user_id"`
	TypeID      int     `json:"type_id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	PhoneNumber string  `json:"phone_number,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	Address     string  `json:"address,omitempty"`
	LogoURL     string  `json:"logo_url,omitempty"`
	CoverURL    string  `json:"cover_url,omitempty"`
}

func ValidateProvider(v *validator.Validator, p *Provider) {
	v.Check(p.Name != "", "name", "must be provided")
	v.Check(len(p.Name) <= 300, "name", "must not be more than 300 bytes long")

	v.Check(p.UserID > 0, "user_id", "must be provided and greater than zero")
	v.Check(p.TypeID > 0, "type_id", "must be provided and greater than zero")

	v.Check(p.Address != "", "address", "must be provided")
	v.Check(len(p.Address) <= 500, "address", "must not be more than 500 bytes long")
}

func (m ProviderModel) Insert(p *Provider) error {
	query := `
		INSERT INTO providers (user_id, name, provider_type_id, address)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`

	args := []any{
		p.UserID,
		p.Name,
		p.TypeID,
		p.Address,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&p.ID)
	if err != nil {
		return fmt.Errorf("inserting provider: %w", err)
	}

	return nil
}

func (m *ProviderModel) Create(p *Provider, ownerName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting tx: %w", err)
	}

	// Rollback on failure
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Insert provider
	query := `
		INSERT INTO providers (user_id, name, provider_type_id, address)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`

	args := []any{
		p.UserID,
		p.Name,
		p.TypeID,
		p.Address,
	}

	err = tx.QueryRowContext(ctx, query, args...).Scan(&p.ID)
	if err != nil {
		return fmt.Errorf("inserting provider: %w", err)
	}

	// Insert staff (owner)
	query = `
		INSERT INTO staff (provider_id, name, is_owner)
		VALUES ($1, $2, true);
	`

	_, err = tx.ExecContext(ctx, query, p.ID, ownerName)
	if err != nil {
		return fmt.Errorf("inserting staff: %w", err)
	}

	return nil
}
