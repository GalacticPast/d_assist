package internal

import "d_assist/internal/auth"

import (
	"net/http"
)

func Verify_cookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := auth.Verify_cookie(r)
		if res != nil {
			http.Redirect(w, r, "/homepage", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
