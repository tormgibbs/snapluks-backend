package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Users                   UserModel
	Providers               ProviderModel
	Tokens                  TokenModel
	EmailVerificationTokens EmailVerificationTokenModel
}

func NewModels(DB *sql.DB) Models {
	return Models{
		Users:                   UserModel{DB},
		Providers:               ProviderModel{DB},
		Tokens:                  TokenModel{DB},
		EmailVerificationTokens: EmailVerificationTokenModel{DB},
	}
}
