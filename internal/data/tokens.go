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
	UserID    int       `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

type TokenModel struct {
	DB *sql.DB
}

// generateToken creates a new token for a user with a given time-to-live (ttl) and scope.
func generateToken(userID int, ttl time.Duration, scope string) (*Token, error) {
	// Create a new token instance with the user ID, expiry time, and scope.
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	// Initialize a zero-valued byte slice with a length of 16 bytes.
	// randomBytes := make([]byte, 16)

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

	// Encode the random bytes as a base32 string without padding for the plaintext token.
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	if scope == ScopeAuthentication {
		token.Plaintext = encoded
	} else {
		token.Plaintext = encoded[:6]
	}

	// Hash the plaintext token using SHA-256 to create a secure, non-reversible version.
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:] // Store the first 32 bytes of the hash in the token's Hash fiel.

	return token, nil
}

// ValidateTokenPlaintext ensures the token is provided and is 26 bytes long.
func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

// New creates a new token for a user with the specified TTL and scope,
// then inserts it into the database.
func (m TokenModel) New(userID int, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

// Insert adds a new token record to the tokens table in the database.
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

// DeleteAllForUser deletes all tokens associated with a particular user
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
		DELETE FROM tokens
		where scope = $1 and user_id = $2
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}
