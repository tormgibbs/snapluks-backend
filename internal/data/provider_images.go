package data

import (
	"context"
	"database/sql"
	"time"
)

type ProviderImageModel struct {
	DB *sql.DB
}

type ProviderImage struct {
	ID         int       `json:"id"`
	ProviderID int64       `json:"provider_id"`
	ImageURL   string    `json:"image_url"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func (m ProviderImageModel) Insert(image *ProviderImage) error {
	query := `
		INSERT INTO provider_images (provider_id, image_url)
		VALUES ($1, $2)
		RETURNING id, uploaded_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, image.ProviderID, image.ImageURL).Scan(
		&image.ID,
		&image.UploadedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m ProviderImageModel) BatchInsert(images []*ProviderImage) error {
	if len(images) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO provider_images (provider_id, image_url)
		VALUES ($1, $2)
		RETURNING id, uploaded_at
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, img := range images {
		err := stmt.QueryRowContext(ctx, img.ProviderID, img.ImageURL).Scan(&img.ID, &img.UploadedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (m ProviderImageModel) GetAllForProvider(providerID int64) ([]*ProviderImage, error) {
	query := `
		SELECT id, provider_id, image_url, uploaded_at
		FROM provider_images
		WHERE provider_id = $1
		ORDER BY uploaded_at DESC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	images := make([]*ProviderImage, 0)

	for rows.Next() {
		var image ProviderImage
		err := rows.Scan(
			&image.ID,
			&image.ProviderID,
			&image.ImageURL,
			&image.UploadedAt,
		)
		if err != nil {
			return nil, err
		}
		images = append(images, &image)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return images, nil
}
