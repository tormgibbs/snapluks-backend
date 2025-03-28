package data

import "github.com/tormgibbs/snapluks-backend/internal/validator"

type Provider struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

func ValidateProvider(v *validator.Validator, p *Provider) {
	v.Check(p.Name != "", "name", "must be provided")
	v.Check(p.Address != "", "address", "must be provided")

	// Latitude validation
	v.Check(p.Latitude != nil, "latitude", "must be provided")
	if p.Latitude != nil {
		v.Check(
			*p.Latitude >= -90 && *p.Latitude <= 90, "latitude", "must be between -90 and 90",
		)
	}
	
	// Longitude validation
	v.Check(p.Longitude != nil, "longitude", "must be provided")
	if p.Longitude != nil {
		v.Check(
			*p.Longitude >= -180 && *p.Longitude <= 180, "longitude", "must be between -180 and 180",
		)
	}
}
