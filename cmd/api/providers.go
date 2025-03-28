package main

import (
	"fmt"
	"net/http"

	"github.com/tormgibbs/snapluks-backend/internal/data"
	"github.com/tormgibbs/snapluks-backend/internal/validator"
)

func (app *application) showProviderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	latitude := 37.85
	longitude := -122.07

	provider := data.Provider{
		ID:        id,
		Name:      "SavingsGrace",
		Address:   "Accra",
		Latitude:  &latitude,
		Longitude: &longitude,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"provider": provider}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createProviderHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name      string   `json:"name"`
		Address   string   `json:"address"`
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	provider := &data.Provider{
		Name: input.Name,
		Address: input.Address,
		Latitude: input.Latitude,
		Longitude: input.Longitude,
	}

	v := validator.New()

	if data.ValidateProvider(v, provider); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// v.Check(input.Name != "", "name", "must be provided")
	// v.Check(input.Address != "", "address", "must be provided")

	// // Latitude validation
	// v.Check(input.Latitude != nil, "latitude", "must be provided")
	// if input.Latitude != nil {
	// 	v.Check(
	// 		*input.Latitude >= -90 && *input.Latitude <= 90, "latitude", "must be between -90 and 90",
	// 	)
	// }
	
	// // Longitude validation
	// v.Check(input.Longitude != nil, "longitude", "must be provided")
	// if input.Longitude != nil {
	// 	v.Check(
	// 		*input.Longitude >= -180 && *input.Longitude <= 180, "longitude", "must be between -180 and 180",
	// 	)
	// }

	// if !v.Valid() {
	// 	app.failedValidationResponse(w, r, v.Errors)
	// 	return
	// }

	fmt.Fprintf(w, "%+v\n", input)
}
