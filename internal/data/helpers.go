package data

// import (
// 	"context"
// 	"fmt"
// 	"time"
// )

// func (m UserModel) InsertClient(u *User, c *Client) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	tx, err := m.DB.BeginTx(ctx, nil)
// 	if err != nil {
// 		return fmt.Errorf("starting tx: %w", err)
// 	}

// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 		} else {
// 			tx.Commit()
// 		}
// 	}()

// 	// Insert user
// 	userQuery := `
// 		INSERT INTO users (email, first_name, last_name, phone_number, password_hash, activated, role)
// 		VALUES ($1, $2, $3, $4, $5, $6, $7)
// 		RETURNING id;
// 	`
// 	err = tx.QueryRowContext(ctx, userQuery,
// 		u.Email,
// 		u.FirstName,
// 		u.LastName,
// 		u.PhoneNumber,
// 		u.Password.hash,
// 		u.Activated,
// 		u.Role,
// 	).Scan(&u.ID)
// 	if err != nil {
// 		return fmt.Errorf("inserting user: %w", err)
// 	}

// 	// Insert client
// 	clientQuery := `INSERT INTO clients (user_id) VALUES ($1) RETURNING id;`
// 	err = tx.QueryRowContext(ctx, clientQuery, u.ID).Scan(&c.ID)
// 	if err != nil {
// 		return fmt.Errorf("inserting client: %w", err)
// 	}

// 	c.UserID = u.ID
// 	return nil
// }

// func (m UserModel) InsertProvider(u *User, p *Provider) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	tx, err := m.DB.BeginTx(ctx, nil)
// 	if err != nil {
// 		return fmt.Errorf("starting tx: %w", err)
// 	}

// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 		} else {
// 			tx.Commit()
// 		}
// 	}()

// 	// Insert user
// 	userQuery := `
// 		INSERT INTO users (email, first_name, last_name, phone_number, password_hash, activated, role)
// 		VALUES ($1, $2, $3, $4, $5, $6, $7)
// 		RETURNING id;
// 	`
// 	err = tx.QueryRowContext(ctx, userQuery,
// 		u.Email,
// 		u.FirstName,
// 		u.LastName,
// 		u.PhoneNumber,
// 		u.Password,
// 		u.Activated,
// 		u.Role,
// 	).Scan(&u.ID)
// 	if err != nil {
// 		return fmt.Errorf("inserting user: %w", err)
// 	}

// 	// Insert provider
// 	providerQuery := `
// 		INSERT INTO providers (user_id, name, type, description, phone_number, location, logo_url, cover_url)
// 		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
// 		RETURNING id;
// 	`
// 	err = tx.QueryRowContext(ctx, providerQuery,
// 		u.ID,
// 		p.Name,
// 		p.Type,
// 		p.Description,
// 		p.PhoneNumber,
// 		p.Location,
// 		p.LogoURL,
// 		p.CoverURL,
// 	).Scan(&p.ID)
// 	if err != nil {
// 		return fmt.Errorf("inserting provider: %w", err)
// 	}

// 	p.UserID = u.ID
// 	return nil
// }
