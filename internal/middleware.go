package internal

import "d_assist/internal/auth"

import (
	"fmt"
	"net/http"
)

func Verify_cookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res, err := auth.Verify_cookie(r)
		if res == false || err != nil {
			http.Error(w, fmt.Sprintf("%v\n", err), http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r)
	})
}
