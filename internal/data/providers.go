package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tormgibbs/snapluks-backend/internal/validator"
)

type ProviderModel struct {
	DB *sql.DB
}

type Provider struct {
	ID          int64   `json:"id"`
	UserID      int64   `json:"user_id"`
	TypeID      int64   `json:"type_id"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
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

	ValidateEmail(v, p.Email)
	ValidatePhone(v, p.PhoneNumber)

	v.Check(p.Description != "", "description", "must be provided")
	v.Check(len(p.Description) <= 10000, "description", "must be not be more than 10000 bytes long")
}

func (m *ProviderModel) Insert(p *Provider, u *User) error {
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

	query := `
		INSERT INTO providers (user_id, name, provider_type_id, phone_number, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`

	args := []any{
		p.UserID,
		p.Name,
		p.TypeID,
		p.PhoneNumber,
		p.Description,
	}

	err = tx.QueryRowContext(ctx, query, args...).Scan(&p.ID)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return ErrDuplicateRecord
		}
		return fmt.Errorf("inserting provider: %w", err)
	}

	// Insert staff (owner)
	query = `
		INSERT INTO staff (provider_id, phone, name, email, is_owner)
		VALUES ($1, $2, $3, $4, true);
	`
	args = []any{
		p.ID,
		u.PhoneNumber,
		u.FirstName,
		u.Email,
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("inserting staff: %w", err)
	}

	return nil
}

func (m ProviderModel) GetByUserID(userID int64) (*Provider, error) {
	query := `
		SELECT id, name, phone_number, description
		FROM providers
		WHERE user_id = $1
	`
	var provider Provider

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, userID).Scan(
		&provider.ID,
		&provider.Name,
		&provider.PhoneNumber,
		&provider.Description,
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
