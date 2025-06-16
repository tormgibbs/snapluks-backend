package main

import (
	"context"
	"net/http"

	"github.com/tormgibbs/snapluks-backend/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}

func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *application) contextGetUserSafe(r *http.Request) (*data.User, bool) {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	return user, ok
}
