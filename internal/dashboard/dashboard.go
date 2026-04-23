package dashboard

import "d_assist/internal/auth"
import "d_assist/internal/db"
import "d_assist/templates/dashboard"

import (
	"github.com/starfederation/datastar-go/datastar"
	"log"
	"net/http"
	"runtime"
)

func Process_upload(w http.ResponseWriter, r *http.Request) {
	res, _ := auth.Get_claims_from_cookie(r)
	if res == false {
		// should we redirect??
		// imma just crash for now
		log.Fatal("This is only possible if the cookie has expired so")
		runtime.Breakpoint()
	}
	component := template_dashboard.Show_spinner()
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(component)
}

func Upload_finished(w http.ResponseWriter, r *http.Request) {
}

func Serve(w http.ResponseWriter, r *http.Request) {
	res, claims := auth.Get_claims_from_cookie(r)
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
