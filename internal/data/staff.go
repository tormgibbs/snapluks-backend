package data

import (
	"context"
	"database/sql"
	"time"
)

type StaffModel struct {
	DB *sql.DB
}

type Staff struct {
	ID             int64  `json:"id"`
	ProviderID     int64  `json:"-"`
	Name           string `json:"name"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	ProfilePicture string `json:"profile_picture"`
	IsOwner        bool   `json:"is_owner"`
}

func (m StaffModel) Insert(s *Staff) error {
	query := `
		INSERT INTO staff (name, phone, email, profile_picture, is_owner, provider_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		s.Name,
		s.Phone,
		s.Email,
		s.ProfilePicture,
		s.IsOwner,
		s.ProviderID,
	}

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&s.ID)

	if err != nil {
		return err
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
