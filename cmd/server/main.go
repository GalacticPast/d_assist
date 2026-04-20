package main

import (
	"d_assist/internal/auth"
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
	err := godotenv.Load("../../.env.local")
	if err != nil {
		log.Fatal("Couldn't load env variables: ", err)
		return
	}
	// init oauth
	auth.Init_oauth()

	frontend_server := http.FileServer(http.Dir("../../static"))
	http.Handle("/", frontend_server)

	// Listen for the Datastar click event
	http.HandleFunc("/auth/google_signup", auth.Google_signup)
	http.HandleFunc("/auth/google_signin", auth.Google_signin)
	http.HandleFunc("/auth/google_callback", auth.Google_callback)

	http.HandleFunc("/interact", serve_interact)
	http.HandleFunc("/validate", validate_login)

	fmt.Println("Server booting up on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func serve_interact(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	// Patches elements into the DOM.
	sse.PatchElements(
		`<div id="response-box">The Go backend says hello!</div>`,
	)
}

func validate_login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("why is this executing? ")
	user_creds := &user_creds{}

	if err := datastar.ReadSignals(r, user_creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}
