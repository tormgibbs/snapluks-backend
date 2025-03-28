package main

import (
	"fmt"
	"net/http"

	"github.com/tormgibbs/snapluks-backend/internal/data"
)

func (app *application) showProviderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	provider := data.Provider{
		ID:        id,
		Name:      "SavingsGrace",
		Address:   "Accra",
		Latitude:  37.85,
		Longitude: -122.07,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"provider": provider}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createProviderHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name      string  `json:"name"`
		Address   string  `json:"address"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	fmt.Fprintf(w, "%+v\n", input)
}
