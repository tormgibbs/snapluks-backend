package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tormgibbs/snapluks-backend/internal/data"
	"github.com/tormgibbs/snapluks-backend/internal/validator"
)

func (app *application) listProvidersHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name      string
		Address   string
		Latitude  float64
		Longitude float64
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Name = app.readString(qs, "name", "")
	input.Address = app.readString(qs, "address", "")
	input.Page = app.readInt(qs, "page", 1, v)
	input.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Sort = app.readString(qs, "sort", "id")

	lat := app.readFloat(qs, "lat", 0, v)
	if lat < -90 || lat > 90 {
		v.AddError("lat", "must be between -90 and 90")
	} else {
		input.Latitude = lat
	}

	long := app.readFloat(qs, "long", 0, v)
	if long < -180 || long > 180 {
		v.AddError("long", "must be between -180 and 180")
	} else {
		input.Longitude = long
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	providers, err := app.models.Providers.GetAll(input.Name, input.Address, input.Latitude, input.Longitude, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": providers}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showProviderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	provider, err := app.models.Providers.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
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
		Name:      input.Name,
		Address:   input.Address,
		Latitude:  input.Latitude,
		Longitude: input.Longitude,
	}

	v := validator.New()

	if data.ValidateProvider(v, provider); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Providers.Insert(provider)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", provider.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"provider": provider}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateProviderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	provider, err := app.models.Providers.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Name      *string  `json:"name"`
		Address   *string  `json:"address"`
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		provider.Name = *input.Name
	}

	if input.Address != nil {
		provider.Address = *input.Address
	}

	if input.Latitude != nil {
		provider.Latitude = input.Latitude
	}

	if input.Longitude != nil {
		provider.Longitude = input.Longitude
	}

	v := validator.New()

	if data.ValidateProvider(v, provider); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Providers.Update(provider)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": provider}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteProviderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Providers.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "provider deleted successfully"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
