package main

import (
	"d_assist/internal/auth"
	"d_assist/internal/dashboard"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/starfederation/datastar-go/datastar"
	"log"
	"net/http"
)

type user_creds struct {
	First_Name string `json:"user_first_name"`
	Last_Name  string `json:"user_last_name"`
	Email      string `json:"user_email"`
	Password   string `json:"user_password"`
}

func main() {
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatal("Couldn't load env variables: ", err)
		return
	}
	// init oauth
	auth.Init_oauth()

	// Listen for the Datastar click event
	static_files := http.Dir("./static")
	fs := http.FileServer(static_files)
	http.Handle("/", http.FileServer(static_files))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/loading", loading_page)
	http.HandleFunc("/dashboard/", dashboard.Serve)

	// auth specific routers
	http.HandleFunc("/auth/google_signup", auth.Google_signup)
	http.HandleFunc("/auth/google_signin", auth.Google_signin)
	http.HandleFunc("/auth/google_callback", auth.Google_callback)

	fmt.Println("Server booting up on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func loading_page(w http.ResponseWriter, r *http.Request) {
	res, _ := auth.Verify_cookie_and_get_claims(r)

	sse := datastar.NewSSE(w, r)
	if res {
		sse.Redirect("/dashboard")
	} else {
		sse.Redirect("/homepage")
	}
}
