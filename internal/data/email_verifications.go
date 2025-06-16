package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"github.com/tormgibbs/snapluks-backend/internal/validator"
)

var (
// ErrRecordNotFound = errors.New("record not found")
// ErrTokenExpired   = errors.New("token has expired")
)

type EmailVerificationTokenModel struct {
	DB *sql.DB
}

type EmailVerificationToken struct {
	Plaintext string    `json:"token"`
	Email     string    `json:"email"`
	Hash      []byte    `json:"-"`
	Expiry    time.Time `json:"expiry"`
}

func (m EmailVerificationTokenModel) New(email string, ttl time.Duration) (*EmailVerificationToken, error) {
	token, err := generateEmailVerificationToken(email, ttl)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

func (m EmailVerificationTokenModel) Insert(token *EmailVerificationToken) error {
	query := `
		INSERT INTO email_verifications (hash, email, expiry)
		VALUES ($1, $2, $3)
	`
	args := []any{token.Hash, token.Email, token.Expiry}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

func generateEmailVerificationToken(email string, ttl time.Duration) (*EmailVerificationToken, error) {
	token := &EmailVerificationToken{
		Email:  email,
		Expiry: time.Now().Add(ttl),
	}

	randomBytes := make([]byte, 16)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

func (m EmailVerificationTokenModel) Verify(tokenPlaintext string) (string, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
		SELECT email
		FROM email_verifications
		WHERE hash = $1 AND expiry > $2
	`
	args := []any{tokenHash[:], time.Now()}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var email string
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&email)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrRecordNotFound
		}
		return "", nil
	}

	return email, nil
}

func (m EmailVerificationTokenModel) DeleteAllForEmail(email string) error {
	query := `
		DELETE FROM email_verifications
		WHERE email = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, email)
	return err
}

func ValidateEmailTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}
