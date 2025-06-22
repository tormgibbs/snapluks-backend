package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tormgibbs/snapluks-backend/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	DB *sql.DB
}

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

type Role string

const (
	RoleClient   Role = "client"
	RoleProvider Role = "provider"
)

type password struct {
	plaintext *string
	hash      []byte
}

type Client struct {
	ID     int `json:"id"`
	UserID int `json:"user_id"`
}

type User struct {
	ID          int            `json:"id"`
	Email       string         `json:"email"`
	FirstName   sql.NullString `json:"first_name,omitempty"`
	LastName    sql.NullString `json:"last_name,omitempty"`
	PhoneNumber sql.NullString `json:"phone_number,omitempty"`
	Password    password       `json:"-"`
	Activated   bool           `json:"activated"`
	Role        Role           `json:"role,omitempty"`
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.FirstName.Valid && user.FirstName.String != "", "first_name", "must be provided")
	v.Check(user.FirstName.Valid || len(user.FirstName.String) <= 500, "first_name", "must not be more than 500 bytes long")

	v.Check(user.LastName.Valid && user.LastName.String != "", "last_name", "must be provided")
	v.Check(!user.LastName.Valid || len(user.LastName.String) <= 500, "last_name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	roles := []string{string(RoleClient), string(RoleProvider)}
	role := string(user.Role)

	v.Check(role != "", "role", "must be provided")
	v.Check(validator.In(role, roles...), "role", "invalid role value")

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

func (m UserModel) Insert(u *User) error {
	query := `
		INSERT INTO users (email, first_name, last_name, phone_number, password_hash, activated, role)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id;
	`

	args := []any{
		u.Email,
		u.FirstName,
		u.LastName,
		u.PhoneNumber,
		u.Password.hash,
		u.Activated,
		u.Role,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&u.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch {
			case pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key":
				return ErrDuplicateEmail
			default:
				return err
			}
		}
	}
	return nil
}

func (m UserModel) InsertInitial(u *User) error {
	query := `
		INSERT INTO users (email)
		VALUES ($1)
		RETURNING id, email, activated;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, u.Email).Scan(&u.ID, &u.Email, &u.Activated)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch {
			case pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key":
				return ErrDuplicateEmail
			default:
				return err
			}
		}
	}
	return nil
}

func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
		SELECT users.id, users.first_name, users.last_name, users.email, users.role, users.password_hash, users.activated
		FROM users
		INNER JOIN tokens
		ON users.id = tokens.user_id
		WHERE tokens.hash = $1
		AND tokens.scope = $2
		AND tokens.expiry > $3
	`
	args := []any{tokenHash[:], tokenScope, time.Now()}

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Role,
		&user.Password.hash,
		&user.Activated,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, email, first_name, last_name, password_hash, activated, role
		FROM users
		where email = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password.hash,
		&user.Activated,
		&user.Role,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) Update(user *User) error {
	query := `
		UPDATE users 
		SET first_name = $1, last_name = $2, phone_number = $3, password_hash = $4, activated = $5, role = $6
		WHERE id = $7
	`
	args := []any{
		user.FirstName,
		user.LastName,
		user.PhoneNumber,
		user.Password.hash,
		user.Activated,
		user.Role,
		user.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	return nil
}
