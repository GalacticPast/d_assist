package auth

import "d_assist/internal/db"
import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/starfederation/datastar-go/datastar"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type auth_config struct {
	google_login_conf oauth2.Config
}

var auth_configs auth_config

func Init_oauth() {

	auth_configs.google_login_conf = oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:8080/auth/google_callback",
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile"},
	}
}

var server_secret = []byte(os.Getenv("GOOGLE_CLIENT_SECRET"))

// gen_rand_string creates the base random payload (e.g., 32 bytes)
func gen_rand_string() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// SignState takes a random string and attaches an HMAC signature to it.
func sign_state(payload string) string {
	// Create a new HMAC using SHA256 and your server's secret key
	mac := hmac.New(sha256.New, server_secret)

	// Write the payload into the hasher
	mac.Write([]byte(payload))

	// Get the resulting signature and encode it to be URL-safe
	signature := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	// Return the format: payload.signature
	return payload + "." + signature
}

// verify_state checks if the state was signed by this server and hasn't been altered.
func verify_state(signedState string) (string, string, error) {
	// Split the incoming state into the payload and the signature
	parts := strings.Split(signedState, ".")
	if len(parts) != 3 {
		return "", "", errors.New("invalid state format")
	}

	payload := parts[0]
	providedSignature := parts[1]
	auth_configs := parts[2]

	// Decode the provided signature from base64
	providedSigBytes, err := base64.URLEncoding.DecodeString(providedSignature)
	if err != nil {
		return "", "", errors.New("invalid signature encoding")
	}

	// Re-sign the payload using our server secret
	mac := hmac.New(sha256.New, server_secret)
	mac.Write([]byte(payload))
	expectedSigBytes := mac.Sum(nil)

	// CRITICAL SECURITY MEASURE: hmac.Equal
	// You MUST use hmac.Equal to compare the signatures, NOT `==`.
	// `==` returns instantly on the first mismatched character, which allows
	// attackers to guess the signature via a "Timing Attack".
	// hmac.Equal always takes the exact same amount of time.
	if !hmac.Equal(providedSigBytes, expectedSigBytes) {
		return "", "", errors.New("signature mismatch: state has been tampered with or forged")
	}

	// It's valid! Return the raw payload.
	return payload, auth_configs, nil
}

func Google_signup(w http.ResponseWriter, r *http.Request) {
	raw_state, _ := gen_rand_string()

	signed_state := sign_state(raw_state)

	signed_state += ".google_signup"

	url := auth_configs.google_login_conf.AuthCodeURL(signed_state)

	sse := datastar.NewSSE(w, r)
	sse.Redirect(url)
}

func Google_signin(w http.ResponseWriter, r *http.Request) {
	raw_state, _ := gen_rand_string()

	signed_state := sign_state(raw_state)
	signed_state += ".google_signin"

	url := auth_configs.google_login_conf.AuthCodeURL(signed_state)

	sse := datastar.NewSSE(w, r)
	sse.Redirect(url)
}

func Google_callback(w http.ResponseWriter, r *http.Request) {
	returned_state := r.FormValue("state")

	_, auth_state, err := verify_state(returned_state)

	if err != nil {
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

	googlecon := auth_configs.google_login_conf

	// 3. Exchange code for token
	token, err := googlecon.Exchange(r.Context(), code)
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

	var user_info db.User_info
	err = json.NewDecoder(resp.Body).Decode(&user_info)
	if err != nil {
		log.Printf("json Unmarshal failed: %v\n", err)
	}

	// we check it anyways
	if auth_state == "google_signin" {
		res := db.Check_if_user_exists(&user_info)

		signed_jwt_token, err := db.Get_JWT_Token(&user_info)

		if err != nil {
			log.Fatalf("Couldn't generate jwt token. %v\n", err)
		}

		auth_cookie := &http.Cookie{
			Name:     "d_assist",
			Value:    signed_jwt_token,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour), // Must match your JWT expiration
			HttpOnly: true,                           // Crucial: Prevents JavaScript from reading the cookie
			Secure:   false,                          // Crucial: Only sends cookie over HTTPS (set to false ONLY on localhost)
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, auth_cookie)

		if res == true {
			http.Redirect(w, r, "/dashboard", http.StatusFound)
		} else {
			http.Redirect(w, r, "/signup", http.StatusUnauthorized)
		}

	} else if auth_state == "google_signup" {
		res := db.Create_user(&user_info)
		if res == true {
			http.Redirect(w, r, "/dashboard", http.StatusFound)
		} else if res == false {
			log.Printf("How is this possible??")
		}
	}

}

func Verify_cookie(cookie *http.Cookie) bool {
	return true
}
