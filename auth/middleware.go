package auth

import (
	"context"
	"errors"
	"net/http"

	e "github.com/ferretcode-freelancing/blog/error"
	"github.com/ferretcode-freelancing/blog/session"
	"gorm.io/gorm"
)

func handleError(w http.ResponseWriter, err error) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func CheckAuth(db *gorm.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("blog")

			if err != nil {
				if errors.Is(http.ErrNoCookie, err) {
					err = e.RenderError(w, r, 403, "You may not access this resource.")
					handleError(w, err)

					return
				}
			}

			sess, err := session.GetSession(cookie.Value)

			if err != nil {
				if errors.Is(session.ErrNotAuthenticated, err) {
					err = e.RenderError(w, r, 403, "You may not access this resource.")
					handleError(w, err)

					return
				} else {
					err = e.RenderError(w, r, 500, "There was an error processing your request.")
					handleError(w, err)

					return
				}
			}

			user := User{}

			err = db.Where("username = ?", sess.Session["username"].(string)).First(&user).Error

			if err != nil {
				err = e.RenderError(w, r, 500, "There was an error processing your request.")
				handleError(w, err)

				return
			}

			ctx := context.WithValue(r.Context(), "user", user)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
