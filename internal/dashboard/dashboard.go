package dashboard

import (
	"d_assist/internal/auth"
	"d_assist/internal/db"
	"d_assist/internal/gemini"
	"d_assist/ui/templates/dashboard"
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
	component := dashboard_templ.Show_spinner()
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(component)
}

func Get_signed_upload_url(w http.ResponseWriter, r *http.Request) {
	file_path := r.FormValue("file_path")
	button_id := r.FormValue("button_id")

	if file_path == "" {
		http.Error(w, "File path is empty", http.StatusBadRequest)
		return
	}
	rand_file_path := rand.Text() + file_path
	signed_url := db.Get_signed_upload_url(button_id, rand_file_path)
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
	button_id := r.FormValue("button_id")
	//@info: shorthand this
	claims := auth.Get_claims_from_cookie(r)
	user_id := auth.Get_user_id_from_claims(&claims)

	pdf_bytes, err := db.Get_pdf_from_bucket(button_id, file_name)
	if err != nil {
		fmt.Errorf("Something wrong with pdf download %v\n", err)
		fmt.Println(err.Error())

		return
	}
	if button_id == "syllabus-upload-btn" {
		syllabus, err := gemini.Extract_syllabus(&pdf_bytes)
		if err != nil {
			err = fmt.Errorf("Something wrong with gemini AI: %v\n", err)
			fmt.Println(err.Error())

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = db.Insert_new_syllabus(user_id, syllabus)
		if err != nil {
			err = fmt.Errorf("Something wrong with supabase: %v\n", err)
			fmt.Println(err.Error())

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if button_id == "slides-upload-btn" {
		badge := r.FormValue("badge")
		deck, err := gemini.Extract_slides(badge, &pdf_bytes)
		if err != nil {
			err = fmt.Errorf("Something wrong with gemini AI: %v\n", err)
			fmt.Println(err.Error())

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = db.Insert_new_deck(user_id, deck)
		if err != nil {
			err = fmt.Errorf("Something wrong with supabase: %v\n", err)
			fmt.Println(err.Error())

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func Serve(w http.ResponseWriter, r *http.Request) {
	claims := auth.Get_claims_from_cookie(r)
	user_id := auth.Get_user_id_from_claims(&claims)
	courses, err := db.Get_courses(user_id)

	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	component := dashboard_templ.Setup(0, courses)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Render the component to the http.ResponseWriter
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, "failed to render template", http.StatusInternalServerError)
	}
}
