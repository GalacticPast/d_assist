package main

import (
	"d_assist/internal/auth"
	"d_assist/internal/dashboard"
	"d_assist/internal/middleware"
)

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/starfederation/datastar-go/datastar"
	"log"
	"net/http"
)

func main() {
	// load env variables
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatal("Couldn't load env variables: ", err)
		return
	}

	// init oauth
	auth.Init_oauth()

	web_fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", web_fs)
	//@perf: there is an extra trip happening
	// idk how to redirect the user when they first connect to the server
	http.HandleFunc("/loading", loading)

	// page specific
	http.Handle("/dashboard", middleware.Verify_cookie(http.HandlerFunc(dashboard.Serve)))

	// auth specific routers
	// auth  need to go through the verify_Cookie handler
	// but there might be different middleware handlers
	http.HandleFunc("/auth/google_signup", auth.Google_signup)
	http.HandleFunc("/auth/google_signin", auth.Google_signin)
	http.HandleFunc("/auth/google_callback", auth.Google_callback)

	// file upload specific
	http.Handle("/process_upload", middleware.Verify_cookie(http.HandlerFunc(dashboard.Process_upload)))
	http.Handle("/upload", middleware.Verify_cookie(http.HandlerFunc(dashboard.Get_signed_upload_url)))
	http.Handle("/upload_finished", middleware.Verify_cookie(http.HandlerFunc(dashboard.Upload_finished)))

	fmt.Println("Server booting up on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func loading(w http.ResponseWriter, r *http.Request) {
	err := auth.Verify_cookie(r)
	sse := datastar.NewSSE(w, r)
	if err != nil {
		sse.Redirect("/homepage")
	} else {
		sse.Redirect("/dashboard")
	}
}
