package main

import (
	"errors"
	"net/http"

	"github.com/tormgibbs/snapluks-backend/internal/data"
	"github.com/tormgibbs/snapluks-backend/internal/validator"
)

func (app *application) createProviderHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	if user.Role != data.RoleProvider {
		app.notPermittedResponse(w, r)
		return
	}

	var input struct {
		Name        string `json:"name"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phone_number"`
		Description string `json:"description"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	provider := &data.Provider{
		TypeID:      1,
		UserID:      user.ID,
		Name:        input.Name,
		Email:       input.Email,
		PhoneNumber: input.PhoneNumber,
		Description: input.Description,
	}

	v := validator.New()
	if data.ValidateProvider(v, provider); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Providers.Insert(provider, user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateRecord):
			v.AddError("provider", "this user already has a provider profile")
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"provider": provider}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateProviderHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
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

	form := r.MultipartForm

	name := app.getFormValue(form, "name")
	email := app.getFormValue(form, "email")
	phone := app.getFormValue(form, "phone_number")
	description := app.getFormValue(form, "description")

	logoFile, logoHeader, err := r.FormFile("logo")
	if err != nil && err != http.ErrMissingFile {
		app.serverErrorResponse(w, r, err)
		return
	}

	coverFile, coverHeader, err := r.FormFile("cover_photo")
	if err != nil && err != http.ErrMissingFile {
		app.serverErrorResponse(w, r, err)
		return
	}

	if name != nil {
		provider.Name = *name
	}
	if email != nil {
		provider.Email = *email
	}
	if phone != nil {
		provider.PhoneNumber = *phone
	}
	if description != nil {
		provider.Description = *description
	}

	if logoFile != nil {
		key, err := app.uploadImageToS3(logoHeader)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		provider.LogoURL = key
	}

	if coverFile != nil {
		key, err := app.uploadImageToS3(coverHeader)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		provider.CoverURL = key
	}

	v := validator.New()

	if data.ValidateProvider(v, provider); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Providers.Update(provider)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"provider": provider}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
