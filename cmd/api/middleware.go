package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/tormgibbs/snapluks-backend/internal/data"
	"github.com/tormgibbs/snapluks-backend/internal/validator"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})

}

func (app *application) requireRole(role data.Role, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.Role != role {
			app.errorResponse(w, r, http.StatusForbidden, "insufficient permissions")
			return
		}

		next(w, r)
	}
}

func (app *application) authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")
		authorizationHeader := r.Header.Get("Authorization")

		// If no Authorization header is present, assign an AnonymousUser to the context
		// and call the next handler. This allows unauthenticated users to proceed.
		// if authorizationHeader == "" {
		// 	r = app.contextSetUser(r, data.AnonymousUser)
		// 	next.ServeHTTP(w, r)
		// 	return
		// }

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		v := validator.New()

		// Validate the token format (this only checks the structure, not whether the token is valid).
		if data.ValidateTokenPlaintext(v, token, data.ScopeAuthentication); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Try to retrieve the user associated with the token from the database.
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// Add the authenticated user to the context so it can be accessed in subsequent handlers.
		r = app.contextSetUser(r, user)

		// Call the next handler in the chain with the updated request.
		next.ServeHTTP(w, r)
	}
}
