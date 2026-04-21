package dashboard

import "d_assist/internal/auth"
import "d_assist/internal/db"
import "d_assist/templates/dashboard"

import (
	//	"github.com/starfederation/datastar-go/datastar"
	"log"
	"net/http"
	"runtime"
)

func Serve(w http.ResponseWriter, r *http.Request) {
	res, claims := auth.Verify_cookie_and_get_claims(r)
	// cookie expired ??
	if res == false {
		// should we redirect??
		// imma just crash for now
		log.Fatal("This is only possible if the cookie has expired so")
		runtime.Breakpoint()
	}
	user_id := auth.Get_user_id_from_claims(&claims)
	count := db.Get_number_of_courses(user_id)

	component := template_dashboard.Setup(count)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Render the component to the http.ResponseWriter
	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, "failed to render template", http.StatusInternalServerError)
	}
}
