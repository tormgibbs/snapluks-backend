package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/api/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/providers/:id", app.showProviderHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/providers", app.createProviderHandler)
	router.HandlerFunc(http.MethodPatch, "/api/v1/providers/:id", app.updateProviderHandler)
	router.HandlerFunc(http.MethodDelete, "/api/v1/providers/:id", app.deleteProviderHandler)

	return router
}
