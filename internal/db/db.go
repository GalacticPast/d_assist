package db

import (
	"github.com/supabase-community/gotrue-go/types"
	"github.com/supabase-community/supabase-go"
	"log"
	"net/http"
	"os"
)

type supabase_client_type int

const (
	SUPABASE_PUBLIC_CLIENT = iota
	SUPABASE_ADMIN_CLIENT  // 1
)

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// @todo: hmmm how to make this one time only, cause for the duration of this app this should only create 2 clients.
// we should just be reusing it right?
func create_supabase_client(client_type supabase_client_type) (*supabase.Client, error) {

	supabase_pub_url := os.Getenv("NEXT_PUBLIC_SUPABASE_URL")

	var client *supabase.Client = nil
	var err error = nil

	switch client_type {
	case SUPABASE_PUBLIC_CLIENT:
		supabase_pub_anon_key := os.Getenv("NEXT_PUBLIC_SUPABASE_ANON_KEY")
		client, err = supabase.NewClient(supabase_pub_url, supabase_pub_anon_key, nil)
	case SUPABASE_ADMIN_CLIENT:
		supabase_priv_secret_key := os.Getenv("NEXT_PRIVATE_SUPABASE_SECRET_KEY")
		client, err = supabase.NewClient(supabase_pub_url, supabase_priv_secret_key, nil)
	}

	if err != nil {
		log.Printf("Failed to initalize the client: %v\n", err)
		return nil, err
	}

	return client, nil
}

func Signin_via_oauth(provider string) string {
	sb_client, err := create_supabase_client(SUPABASE_PUBLIC_CLIENT)
	if err != nil {
		return ""
	}

	res, err := sb_client.Auth.Authorize(types.AuthorizeRequest{
		Provider: types.Provider(provider),
		FlowType: types.FlowType("https://wtpfmvqjwzkwtsvswtmm.supabase.co/auth/v1/callback"),
	})

	if err != nil {
		log.Printf("Failed to authorize the user : %v\n", err)
		return ""
	}

	return res.AuthorizationURL
}

func signin_callback(w http.ResponseWriter, r *http.Request) {
	log.Println("Google auth callback susssfullly!!!!")
}

func Check_if_user_exists(user_data User) bool {

	return false
}
