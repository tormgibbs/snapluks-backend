package main

import (
	"errors"
	"mime/multipart"
	"net/http"
	"sync"

	"github.com/tormgibbs/snapluks-backend/internal/data"
	"github.com/tormgibbs/snapluks-backend/internal/validator"
)

func (app *application) createServiceHandler(w http.ResponseWriter, r *http.Request) {
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
		Name        string                  `form:"name"`
		Description string                  `form:"description"`
		Duration    string                  `form:"duration"`
		Price       float64                 `form:"price"`
		TypeID      int32                   `form:"type_id"`
		CategoryIDs []int32                 `form:"categories"`
		StaffIDs    []int64                 `form:"staff"`
		Images      []*multipart.FileHeader `form:"images"`
	}

	err = app.readMultipartForm(r, 10<<20, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	service := &data.Service{
		Name:        input.Name,
		ProviderID:  provider.ID,
		TypeID:      input.TypeID,
		Categories:  input.CategoryIDs,
		Description: input.Description,
		Duration:    input.Duration,
		Price:       input.Price,
		Staff:       input.StaffIDs,
	}

	v := validator.New()

	if data.ValidateService(v, service); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	if len(input.Images) > 5 {
		v.AddError("images", "cannot upload more than 5 images")
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

	images := input.Images

	uploadedImages := make([]string, len(images))
	var wg sync.WaitGroup
	errChan := make(chan error, len(images))

	for i, fileHeader := range images {
		wg.Add(1)
		go func(index int, fh *multipart.FileHeader) {
			defer wg.Done()

			key, err := app.uploadImageToS3(fh)
			if err != nil {
				errChan <- err
				return
			}
			uploadedImages[index] = key

			err = app.models.Services.InsertImage(service.ID, provider.ID, key, index == 0)
			if err != nil {
				errChan <- err
				return
			}
		}(i, fileHeader)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		app.cleanupFailedServiceCreation(service.ID, provider.ID, uploadedImages)
		app.serverErrorResponse(w, r, <-errChan)
		return
	}

	service.Images = uploadedImages

	err = app.writeJSON(w, http.StatusCreated, envelope{"service": service}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listServiceHandler(w http.ResponseWriter, r *http.Request) {
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

	services, err := app.models.Services.GetAllForProvider(provider.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"services": services}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
