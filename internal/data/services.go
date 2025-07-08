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
	ErrDuplicateService = errors.New("service already exists")
	ErrCategoryNotFound = errors.New("category not found")
	ErrStaffNotFound    = errors.New("staff not found")
)

type ServiceModel struct {
	DB *sql.DB
}

type Service struct {
	ID          int64    `json:"id"`
	ProviderID  int64    `json:"-"`
	TypeID      int32    `json:"type_id"`
	Categories  []int32  `json:"categories"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Duration    string   `json:"duration"`
	Price       float64  `json:"price"`
	Staff       []int64  `json:"staff"`
	Images      []string `json:"images"`
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
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return ErrDuplicateRecord
		}
		return err
	}

	query = `
		INSERT INTO service_categories (service_id, category_id, provider_id)
		VALUES ($1, $2, $3)
	`
	for _, categoryID := range s.Categories {
		_, err = tx.ExecContext(ctx, query, s.ID, categoryID, s.ProviderID)
		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" {
				return ErrCategoryNotFound
			}
			return err
		}
	}

	query = `
		INSERT INTO staff_services (staff_id, service_id, provider_id)
		VALUES ($1, $2, $3)
	`
	for _, staffID := range s.Staff {
		_, err = tx.ExecContext(ctx, query, staffID, s.ID, s.ProviderID)
		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" {
				return ErrStaffNotFound
			}
			return err
		}
	}

	return nil
}

func (m ServiceModel) GetAllForProvider(providerID int64) ([]*Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT s.id, s.name, s.description, s.duration, s.price, s.type_id, s.provider_id,
			COALESCE(array_agg(DISTINCT c.id) FILTER (WHERE c.id IS NOT NULL), '{}') as category_ids,
			COALESCE(array_agg(DISTINCT st.id) FILTER (WHERE st.id IS NOT NULL), '{}') as staff_ids
		FROM services s
		LEFT JOIN service_categories sc ON s.id = sc.service_id
		LEFT JOIN categories c ON sc.category_id = c.id
		LEFT JOIN staff_services ss ON s.id = ss.service_id
		LEFT JOIN staff st ON ss.staff_id = st.id
		WHERE s.provider_id = $1
		GROUP BY s.id, s.name, s.description, s.duration, s.price, s.type_id, s.provider_id
		ORDER BY s.name
	`

	rows, err := m.DB.QueryContext(ctx, query, providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := make([]*Service, 0)

	for rows.Next() {
		var service Service
		var categoryIDs []int32
		var staffIDs []int64

		err := rows.Scan(
			&service.ID,
			&service.Name,
			&service.Description,
			&service.Duration,
			&service.Price,
			&service.TypeID,
			&service.ProviderID,
			&categoryIDs,
			&staffIDs,
		)
		if err != nil {
			return nil, err
		}

		service.Categories = make([]int32, len(categoryIDs))
		copy(service.Categories, categoryIDs)

		service.Staff = make([]int64, len(staffIDs))
		copy(service.Staff, staffIDs)

		services = append(services, &service)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}

func (m ServiceModel) InsertImage(serviceID, providerID int64, imageURL string, isPrimary bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO service_images (service_id, provider_id, image_url, is_primary)
		VALUES ($1, $2, $3, $4)
	`
	_, err := m.DB.ExecContext(ctx, query, serviceID, providerID, imageURL, isPrimary)
	return err
}

func (m ServiceModel) InsertImages(serviceID, providerID int64, imageKeys []string) error {
	if len(imageKeys) == 0 {
		return nil
	}

	query := `
		INSERT INTO service_images (service_id, provider_id, image_url, is_primary)
		VALUES 
	`
	args := make([]any, 0, len(imageKeys)*4)
	placeholders := make([]string, 0, len(imageKeys))

	for i, key := range imageKeys {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d)",
			i*4+1, i*4+2, i*4+3, i*4+4))
		args = append(args, serviceID, providerID, key, i == 0) // First image is primary
	}

	query += strings.Join(placeholders, ", ")

	_, err := m.DB.Exec(query, args...)
	return err
}

func (m ServiceModel) Delete(serviceID, providerID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		DELETE FROM services 
		WHERE id = $1 AND provider_id = $2
	`
	result, err := m.DB.ExecContext(ctx, query, serviceID, providerID)
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
