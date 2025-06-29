package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tormgibbs/snapluks-backend/internal/validator"
)

var ErrDuplicateCategory = errors.New("category already exists")

type CategoryModel struct {
	DB *sql.DB
}

type Category struct {
	ID         int32  `json:"id"`
	ProviderID int64  `json:"-"`
	Name       string `json:"category"`
}

func ValidateCategory(v *validator.Validator, category string) {
	v.Check(category != "", "category", "must be provided")
	v.Check(validator.Matches(category, validator.CategoryRX), "category", "must only contain letters, numbers, and spaces")
	v.Check(len(category) >= 3 && len(category) <= 50, "category", "must be between 3 and 50 bytes")
}

func (m CategoryModel) Insert(c *Category) error {
	query := `
		INSERT INTO categories (provider_id, name)
		VALUES ($1, $2)
		RETURNING id;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, c.ProviderID, c.Name).Scan(&c.ID)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				return ErrDuplicateCategory
			}
		}
		return err
	}

	return nil
}

func (m CategoryModel) GetAllByProviderID(providerID int64) ([]*Category, error) {
	query := `
		SELECT id, name
		FROM categories
		WHERE provider_id = $1
		ORDER BY name
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*Category

	for rows.Next() {
		var c Category
		err := rows.Scan(&c.ID, &c.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}
