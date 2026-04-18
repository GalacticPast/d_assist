package auth

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/starfederation/datastar-go/datastar"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"net/http"
	"os"
)

type state struct {
	google_login_conf oauth2.Config
}

var auth_state state

func Init_oauth() {
	//load .env file
	err := godotenv.Load(".env.local")
	if err != nil {
		fmt.Println("wtf")
	}
	//g_client_id := os.Getenv("GOOGLE_CLIENT_ID")

	auth_state.google_login_conf = oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:8080/dashboard",
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile"},
	}
}

func Google_login(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)
	url := auth_state.google_login_conf.AuthCodeURL("randomstate")
	sse.Redirect(url)
}
