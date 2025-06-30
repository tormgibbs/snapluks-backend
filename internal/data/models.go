package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	ErrDuplicateRecord = errors.New("duplicate record")
)

type Models struct {
	Users                   UserModel
	Providers               ProviderModel
	Tokens                  TokenModel
	Categories              CategoryModel
	Services                ServiceModel
	Staff                   StaffModel
	EmailVerificationTokens EmailVerificationTokenModel
}

func NewModels(DB *sql.DB) Models {
	return Models{
		Users:                   UserModel{DB},
		Providers:               ProviderModel{DB},
		Tokens:                  TokenModel{DB},
		Services:                ServiceModel{DB},
		Staff:                   StaffModel{DB},
		Categories:              CategoryModel{DB},
		EmailVerificationTokens: EmailVerificationTokenModel{DB},
	}
}
