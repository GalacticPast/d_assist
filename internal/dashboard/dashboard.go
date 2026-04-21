package dashboard

import "d_assist/internal/auth"
import "d_assist/internal/db"

import (
	"github.com/starfederation/datastar-go/datastar"
	"log"
	"net/http"
	"runtime"
)

func Setup(w http.ResponseWriter, r *http.Request) {
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
	sse := datastar.NewSSE(w, r)

	if count == 0 {
		sse.PatchElements(
			`<div class="empty_shit"> Empty courses </div>`,
		)
	} else {

	}

}
