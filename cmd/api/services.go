package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/tormgibbs/snapluks-backend/internal/data"
	"github.com/tormgibbs/snapluks-backend/internal/validator"
)

func (app *application) createServiceHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	description := strings.TrimSpace(r.FormValue("description"))
	duration := strings.TrimSpace(r.FormValue("duration"))
	priceStr := strings.TrimSpace(r.FormValue("price"))
	typeIDStr := strings.TrimSpace(r.FormValue("type_id"))
	categories := r.Form["categories"]
	staff := r.Form["staff"]

	v := validator.New()

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		v.AddError("price", "invalid price format")
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	typeID, err := strconv.Atoi(typeIDStr)
	if err != nil {
		v.AddError("type_id", "invalid type_id format")
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	categoryIDs, err := data.ParseIntSlice[int32](categories)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	staffIDs, err := data.ParseIntSlice[int64](staff)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	files := r.MultipartForm.File["images"]
	if len(files) > 5 {
		v.AddError("images", "cannot upload more than 5 images")
		app.failedValidationResponse(w, r, v.Errors)
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
		Name:        name,
		ProviderID:  provider.ID,
		TypeID:      int32(typeID),
		Categories:  categoryIDs,
		Description: description,
		Duration:    duration,
		Price:       price,
		Staff:       staffIDs,
	}

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

	uploadedImages := make([]string, 0)
	for i, fileHeader := range files {
		key, err := app.uploadImageToS3(fileHeader)
		if err != nil {
			app.cleanupFailedServiceCreation(service.ID, provider.ID, uploadedImages)
			app.serverErrorResponse(w, r, err)
			return
		}
		uploadedImages = append(uploadedImages, key)

		err = app.models.Services.InsertImage(service.ID, provider.ID, key, i == 0)
		if err != nil {
			app.cleanupFailedServiceCreation(service.ID, provider.ID, uploadedImages)
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	service.Images = uploadedImages

	err = app.writeJSON(w, http.StatusCreated, envelope{"service": service}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
