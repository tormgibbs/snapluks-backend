package main

import (
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/tormgibbs/snapluks-backend/internal/data"
	"github.com/tormgibbs/snapluks-backend/internal/validator"
)

func (app *application) createStaffHandler(w http.ResponseWriter, r *http.Request) {
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

	var input struct {
		Name           string                `form:"name"`
		Phone          string                `form:"phone"`
		Email          string                `form:"email"`
		Services       []int64               `form:"services"`
		ProfilePicture *multipart.FileHeader `form:"profile_picture"`
	}

	err = app.readMultipartForm(r, 10<<20, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	staff := &data.Staff{
		ProviderID:     provider.ID,
		Name:           input.Name,
		Phone:          input.Phone,
		Email:          input.Email,
		Services:       input.Services,
	}

	if input.ProfilePicture == nil {
		staff.ProfilePicture = nil
	}

	v := validator.New()

	if data.ValidateStaff(v, staff); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Staff.Insert(staff)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrServiceNotFound):
			v.AddError("services", "one or more services not found")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"staff": staff}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listStaffHandler(w http.ResponseWriter, r *http.Request) {
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

	staff, err := app.models.Staff.GetAllForProvider(provider.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"staff": staff}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
