package middleware

import (
	"net/http"
	"strings"

	"github.com/arnoldokoth/lenslocked.com/context"
	"github.com/arnoldokoth/lenslocked.com/models"
)

// User ...
type User struct {
	models.UserService
}

// Apply ...
func (mw *User) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}

// ApplyFn ...
func (mw *User) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/assets/") || strings.HasPrefix(r.URL.Path, "/images/") {
			next(w, r)
			return
		}
		cookie, err := r.Cookie("remember_token")
		if err != nil {
			next(w, r)
			return
		}

		user, err := mw.UserService.ByRemember(cookie.Value)
		if err != nil {
			next(w, r)
			return
		}

		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		r = r.WithContext(ctx)

		next(w, r)
	})
}

// RequireUser ...
type RequireUser struct {
	User
}

// Apply ...
func (mw *RequireUser) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}

// ApplyFn ...
func (mw *RequireUser) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return mw.User.Apply(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w, r)
	}))
}
