package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodPost, "/api/v1/auth/request-verification", app.requestVerificationHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/auth/verify-token", app.verifyTokenHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/auth/login", app.loginHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/auth/register", app.registerUserHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/providers", app.authenticate(app.createProviderHandler))


	return router
}
