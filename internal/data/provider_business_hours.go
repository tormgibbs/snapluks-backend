package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/tormgibbs/snapluks-backend/internal/validator"
)

type ProviderBusinessHoursModel struct {
	DB *sql.DB
}

type ProviderBusinessHour struct {
	ID         int        `json:"id"`
	ProviderID int        `json:"provider_id"`
	DayOfWeek  int        `json:"day_of_week"`
	IsClosed   bool       `json:"is_closed"`
	OpenTime   *LocalTime `json:"open_time,omitempty"`
	CloseTime  *LocalTime `json:"close_time,omitempty"`
}

func ValidateProviderBusinessHour(v *validator.Validator, bh *ProviderBusinessHour) {
	v.Check(bh.ProviderID > 0, "provider_id", "must be provided and greater than zero")
	v.Check(bh.DayOfWeek >= 0 && bh.DayOfWeek <= 6, "day_of_week", "must be between 0 (Sunday) and 6 (Saturday)")

	if bh.IsClosed {
		v.Check(bh.OpenTime == nil, "open_time", "must be null when is_closed is true")
		v.Check(bh.CloseTime == nil, "close_time", "must be null when is_closed is true")
	} else {
		v.Check(bh.OpenTime != nil, "open_time", "must be provided when is_closed is false")
		v.Check(bh.CloseTime != nil, "close_time", "must be provided when is_closed is false")

		if bh.OpenTime != nil && bh.CloseTime != nil {
			v.Check(bh.OpenTime.Before(*bh.CloseTime), "time", "open_time must be before close_time")
		}
	}
}

func (m *ProviderBusinessHoursModel) Insert(bh *ProviderBusinessHour) error {
	query := `
		INSERT INTO provider_business_hours (provider_id, day_of_week, is_closed, open_time, close_time)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(
		ctx,
		query,
		bh.ProviderID,
		bh.DayOfWeek,
		bh.IsClosed,
		bh.OpenTime,
		bh.CloseTime,
	).Scan(&bh.ID)
}

func (m *ProviderBusinessHoursModel) GetAllForProvider(providerID int64) ([]*ProviderBusinessHour, error) {
	query := `
		SELECT id, provider_id, day_of_week, is_closed, open_time, close_time
		FROM provider_business_hours
		WHERE provider_id = $1
		ORDER BY day_of_week
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hours []*ProviderBusinessHour

	for rows.Next() {
		var bh ProviderBusinessHour
		err := rows.Scan(
			&bh.ID,
			&bh.ProviderID,
			&bh.DayOfWeek,
			&bh.IsClosed,
			&bh.OpenTime,
			&bh.CloseTime,
		)
		if err != nil {
			return nil, err
		}
		hours = append(hours, &bh)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return hours, nil
}

func (m *ProviderBusinessHoursModel) Update(ctx context.Context, bh *ProviderBusinessHour) error {
	query := `
		UPDATE provider_business_hours
		SET is_closed = $1,
		    open_time = $2,
		    close_time = $3
		WHERE id = $4
	`

	_, err := m.DB.ExecContext(
		ctx,
		query,
		bh.IsClosed,
		bh.OpenTime,
		bh.CloseTime,
		bh.ID,
	)

	return err
}
