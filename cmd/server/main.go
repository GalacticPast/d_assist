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

	frontend_server := http.FileServer(http.Dir("./static"))
	http.Handle("/", frontend_server)

	// Listen for the Datastar click event
	http.HandleFunc("/loading", loading_page)

	// auth specific routers

	http.HandleFunc("/auth/google_signup", auth.Google_signup)
	http.HandleFunc("/auth/google_signin", auth.Google_signin)
	http.HandleFunc("/auth/google_callback", auth.Google_callback)

	http.HandleFunc("/dashboard_setup", dashboard.Setup)

	fmt.Println("Server booting up on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func loading_page(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	cookie := r.CookiesNamed("d_assist")
	if len(cookie) == 0 {
		fmt.Print("No cookie found")
		sse.Redirect("/homepage")
		return
	}
	res := auth.Verify_cookie(cookie[0])

	if res == true {
		sse.Redirect("/dashboard")
	} else {
		sse.Redirect("/homepage")
	}
}
