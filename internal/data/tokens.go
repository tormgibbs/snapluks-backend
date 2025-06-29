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

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
	ScopePasswordReset  = "password-reset"
)

// Token struct represents the structure of a token.
type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64       `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

type TokenModel struct {
	DB *sql.DB
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	var randomBytes []byte

	switch scope {
	case ScopeAuthentication:
		randomBytes = make([]byte, 16)
	default:
		randomBytes = make([]byte, 4)
	}

	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	if scope == ScopeAuthentication {
		token.Plaintext = encoded
	} else {
		token.Plaintext = encoded[:6]
	}

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string, scope string) {

	v.Check(tokenPlaintext != "", "token", "must be provided")

	switch scope {
	case ScopeAuthentication:
		v.Check(len(tokenPlaintext) == 26, "token", "must be 26 characters long")
	default:
		v.Check(len(tokenPlaintext) == 6, "token", "must be 6 characters long")
	}
}

func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

func (m TokenModel) Insert(token *Token) error {
	query := `
		INSERT INTO tokens (hash, user_id, expiry, scope)
		VALUES ($1, $2, $3, $4)
	`

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

func (m TokenModel) DeleteAllForUser(userID int64, scope string) error {
	query := `
		DELETE FROM tokens
		where scope = $1 and user_id = $2
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}
