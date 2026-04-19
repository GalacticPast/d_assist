package auth

import (
	"d_assist/internal/db"
	"github.com/starfederation/datastar-go/datastar"
	"net/http"
)

func Google_signin(w http.ResponseWriter, r *http.Request) {
	url := db.Signin_via_oauth("google")

	sse := datastar.NewSSE(w, r)
	sse.Redirect(url)
}

func Callback(w http.ResponseWriter, r *http.Request) {

}
