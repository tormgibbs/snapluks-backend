package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tormgibbs/snapluks-backend/internal/validator"
)

var (
	ErrServiceNotFound = errors.New("service not found")
)

type StaffModel struct {
	DB *sql.DB
}

type Staff struct {
	ID             int64   `json:"id"`
	ProviderID     int64   `json:"-"`
	Name           string  `json:"name"`
	Phone          string  `json:"phone"`
	Email          string  `json:"email"`
	ProfilePicture string  `json:"profile_picture"`
	IsOwner        bool    `json:"is_owner"`
	Services       []int64 `json:"services"`
}

func ValidateStaff(v *validator.Validator, s *Staff) {
	// Name: required, 1–100 characters
	v.Check(s.Name != "", "name", "must be provided")
	v.Check(len(s.Name) <= 100, "name", "must not be more than 100 characters")

	// Phone: required, basic format check (adjust regex as needed)
	v.Check(s.Phone != "", "phone", "must be provided")
	v.Check(validator.Matches(s.Phone, validator.PhoneRX), "phone", "must be a valid phone number")

	// Email: required, format check
	v.Check(s.Email != "", "email", "must be provided")
	v.Check(validator.Matches(s.Email, validator.EmailRX), "email", "must be a valid email address")

	// Services: optional, but if provided, ensure all IDs are positive
	v.Check(len(s.Services) > 0, "services", "at least one service must be provided")
	for i, id := range s.Services {
		v.Check(id > 0, fmt.Sprintf("services[%d]", i), "must be a valid service ID")
	}
}

func (m StaffModel) Insert(s *Staff) error {
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
		INSERT INTO staff (name, phone, email, profile_picture, is_owner, provider_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	args := []any{
		s.Name,
		s.Phone,
		s.Email,
		s.ProfilePicture,
		s.IsOwner,
		s.ProviderID,
	}

	err = tx.QueryRowContext(ctx, query, args...).Scan(&s.ID)
	if err != nil {
		return err
	}

	if len(s.Services) > 0 {
		placeholders := make([]string, len(s.Services))
		args := make([]any, 0, len(s.Services)+2)

		args = append(args, s.ID, s.ProviderID)

		for i, serviceID := range s.Services {
			placeholders[i] = fmt.Sprintf("($1, $%d, $2)", i+3)
			args = append(args, serviceID)
		}

		query = fmt.Sprintf(`
			INSERT INTO staff_services (staff_id, service_id, provider_id)
			VALUES %s
		`, strings.Join(placeholders, ", "))

		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" {
				return ErrServiceNotFound
			}
			return err
		}
	}

	return nil
}

func (m StaffModel) GetAllByProviderID(providerID int64) ([]*Staff, error) {
	query := `
		SELECT id, name, phone, email, profile_picture, is_owner
		FROM staff
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

	var staffList []*Staff

	for rows.Next() {
		var s Staff
		err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.Phone,
			&s.Email,
			&s.ProfilePicture,
			&s.IsOwner,
		)
		if err != nil {
			return nil, err
		}

		staffList = append(staffList, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return staffList, nil
}
