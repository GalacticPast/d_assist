package auth

import "d_assist/internal/db"

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"github.com/starfederation/datastar-go/datastar"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io"
	"log"
	"net/http"
	"os"
)

type state struct {
	google_login_conf oauth2.Config
}

var auth_state state

func Init_oauth() {
	//load .env file
	err := godotenv.Load("./.env.local")
	if err != nil {
		log.Fatal("Error: ", err)
		return
	}
	//g_client_id := os.Getenv("GOOGLE_CLIENT_ID")

	auth_state.google_login_conf = oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:8080/google_callback",
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile"},
	}
}

func Google_login(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)
	url := auth_state.google_login_conf.AuthCodeURL("randomstate")
	sse.Redirect(url)
}

func Google_callback(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != "randomstate" {
		log.Println("States don't match")
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		log.Println("Code not found in request")
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	google_config := auth_state.google_login_conf

	token, err := google_config.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("Code-Token Exchange Failed: %v\n", err)
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	// 4. Fetch User Data
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		log.Printf("User Data Fetch Failed: %v\n", err)
		http.Error(w, "Failed to fetch user data", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	google_user_data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("JSON Parsing Failed: %v\n", err)
		http.Error(w, "Failed to read user data", http.StatusInternalServerError)
		return
	}

	user := db.User{}
	err = json.Unmarshal(google_user_data, &user)
	if err != nil {
		log.Printf("JSON Parsing Failed: %v\n", err)
		http.Error(w, "Failed to Unmarshal user data", http.StatusInternalServerError)
		return
	}

	//@todo: finish this
	res := db.Check_if_user_exists(user)

	if res == false {
		http.Redirect(w, r, "/login", http.StatusNotFound)
	} else {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	}

}
