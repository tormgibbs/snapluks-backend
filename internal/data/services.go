package data

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/tormgibbs/snapluks-backend/internal/validator"
)

type ServiceModel struct {
	DB *sql.DB
}

type Service struct {
	ID          int64   `json:"id"`
	ProviderID  int64   `json:"-"`
	TypeID      int32   `json:"type_id"`
	Categories  []int32 `json:"categories"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Duration    string  `json:"duration"`
	Price       float64 `json:"price"`
	Staff       []int64 `json:"staff"`
}

func validateDuration(v *validator.Validator, duration string) {
	if duration == "" {
		v.AddError("duration", "must be provided")
		return
	}
	if _, err := time.ParseDuration(duration); err != nil {
		v.AddError("duration", "must be a valid duration (e.g. '30m', '1h')")
	}
}

func ValidateService(v *validator.Validator, s *Service) {
	v.Check(strings.TrimSpace(s.Name) != "", "name", "must be provided")
	v.Check(strings.TrimSpace(s.Description) != "", "description", "must be provided")
	v.Check(s.Price > 0, "price", "must be greater than zero")
	v.Check(s.TypeID != 0, "type_id", "must be provided")

	validateDuration(v, s.Duration)

	validateIDSlice(v, (s.Categories), "categories")
	validateIDSlice(v, s.Staff, "staff")
}

func validateIDSlice[T ~int | ~int32 | ~int64](v *validator.Validator, values []T, field string) {
	v.Check(len(values) > 0, field, fmt.Sprintf("must include at least one %s", field))
	if validator.HasDuplicates(values) {
		v.AddError(field, "must not contain duplicate values")
	}
	for _, id := range values {
		if id <= 0 {
			v.AddError(field, fmt.Sprintf("contains invalid %s ID", field))
			break
		}
	}
}

func (m ServiceModel) Insert(s *Service) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

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
		INSERT INTO services (name, description, duration, price, type_id, provider_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	args := []any{
		s.Name,
		s.Description,
		s.Duration,
		s.Price,
		s.TypeID,
		s.ProviderID,
	}

	err = tx.QueryRowContext(ctx, query, args...).Scan(&s.ID)
	if err != nil {
		return err
	}

	query = `
		INSERT INTO service_categories (service_id, category_id)
		VALUES ($1, $2)
	`
	for _, categoryID := range s.Categories {
		_, err = tx.ExecContext(ctx, query, s.ID, categoryID)
		if err != nil {
			return err
		}
	}

	query = `
		INSERT INTO staff_services (staff_id, service_id)
		VALUES ($1, $2)
	`
	for _, staffID := range s.Staff {
		_, err = tx.ExecContext(ctx, query, staffID, s.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
