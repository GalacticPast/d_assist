package dashboard

import (
	"d_assist/internal/auth"
	"d_assist/internal/db"
	"d_assist/internal/gemini"
	"d_assist/templates/dashboard"
)

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/starfederation/datastar-go/datastar"
	"net/http"
)

type Signed_url struct {
	File_name string `json:"file_name"`
	URL       string `json:"url"`
}

func Process_upload(w http.ResponseWriter, r *http.Request) {
	component := template_dashboard.Show_spinner()
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(component)
}

func Get_signed_upload_url(w http.ResponseWriter, r *http.Request) {
	file_path := r.FormValue("file_path")
	if file_path == "" {
		http.Error(w, "File path is empty", http.StatusBadRequest)
		return
	}
	rand_file_path := rand.Text() + file_path
	signed_url := db.Get_signed_upload_url(rand_file_path)
	response := Signed_url{
		File_name: rand_file_path,
		URL:       signed_url,
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func Upload_finished(w http.ResponseWriter, r *http.Request) {
	file_name := r.FormValue("file_path")
	pdf_bytes, err := db.Get_pdf_from_bucket(file_name)
	if err != nil {
		fmt.Errorf("Something wrong with pdf download %v\n", err)
		return
	}
	gemini.Extract_courses(&pdf_bytes)
}

func Serve(w http.ResponseWriter, r *http.Request) {
	claims := auth.Get_claims_from_cookie(r)
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
