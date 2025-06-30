package main

import (
	"errors"
	"net/http"

	"github.com/tormgibbs/snapluks-backend/internal/data"
	"github.com/tormgibbs/snapluks-backend/internal/validator"
)

func (app *application) createServiceHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TypeID      int32   `json:"type_id"`
		Categories  []int32 `json:"categories"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Duration    string  `json:"duration"`
		Price       float64 `json:"price"`
		Staff       []int64 `json:"staff"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	if user.Role != data.RoleProvider {
		app.notPermittedResponse(w, r)
		return
	}

	provider, err := app.models.Providers.GetByUserID(user.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			msg := "you must setup a provider profile"
			app.notPermittedWithMessageResponse(w, r, msg)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	service := &data.Service{
		Name:        input.Name,
		ProviderID:  provider.ID,
		TypeID:      input.TypeID,
		Categories:  input.Categories,
		Description: input.Description,
		Duration:    input.Duration,
		Price:       input.Price,
		Staff:       input.Staff,
	}

	v := validator.New()

	if data.ValidateService(v, service); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Services.Insert(service)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateRecord):
			v.AddError("service", "a service with that name already exists for this provider")
			app.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, data.ErrCategoryNotFound):
			v.AddError("category", "one or more provided categories were not found")
			app.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, data.ErrStaffNotFound):
			v.AddError("staff", "one or more selected staff members do not exist")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"service": service}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
