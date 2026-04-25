package homepage

import "d_assist/ui/templates/homepage"
import "net/http"

func Serve(w http.ResponseWriter, r *http.Request) {
	component := homepage_templ.Setup()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Render the component to the http.ResponseWriter
	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, "failed to render template", http.StatusInternalServerError)
	}
}
