package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodPost, "/api/v1/auth/register", app.registerUserHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/auth/resend-verification", app.resendVerificationHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/auth/verify-email", app.verifyEmailHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/auth/complete-profile", app.completeProfileHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/auth/login", app.loginHandler)

	router.HandlerFunc(http.MethodPost, "/api/v1/providers", app.authenticate(app.createProviderHandler))

	router.HandlerFunc(http.MethodPost, "/api/v1/categories", app.authenticate(app.createCategoryHandler))

	router.HandlerFunc(http.MethodPost, "/api/v1/services", app.authenticate(app.createServiceHandler))

	return router
}
