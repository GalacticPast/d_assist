package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/supabase-community/gotrue-go/types"
	"github.com/supabase-community/supabase-go"
	"log"
	"os"
)

type supabase_client_type int

const (
	SUPABASE_PUBLIC_CLIENT = iota
	SUPABASE_ADMIN_CLIENT  // 1
)

type User_info struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// @todo: hmmm how to make this one time only, cause for the duration of this app this should only create 2 clients.
// we should just be reusing it right?
func Create_supabase_client(client_type supabase_client_type) (*supabase.Client, error) {

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

func Exchange_code(code string) types.Session {
	supabase_client, err := Create_supabase_client(SUPABASE_PUBLIC_CLIENT)
	if err != nil {

	}

	token_response, err := supabase_client.Auth.Token(types.TokenRequest{
		GrantType: "pcke",
		Code:      code,
	})

	if err != nil {
		log.Fatalf("Couldnt authorize user.")
	}

	return token_response.Session
}

func Signin_via_oauth(provider string) string {
	sb_client, err := Create_supabase_client(SUPABASE_PUBLIC_CLIENT)
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

func Check_if_user_exists(user_data *User_info) bool {
	// @todo: have a connection pool
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer conn.Close(context.Background())

	// Example query to test connection
	// 1. Use $1 as the placeholder for your variable
	query := "SELECT EXISTS(SELECT 1 FROM profiles WHERE id = $1)"

	// 2. Pass the variable as a separate argument to QueryRow
	var exists bool
	err = conn.QueryRow(context.Background(), query, user_data.ID).Scan(&exists)

	if err != nil {
		// Handle the database error (e.g., connection lost)
		log.Fatalf("Query failed: %v\n", err)
	}
	if exists {
		return true
	}
	return false
}

func Create_user(user_data *User_info) bool {
	// @todo: have a connection pool
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer conn.Close(context.Background())

	query := "INSERT INTO profiles (id, name, email) VALUES (@id, @name, @email)"
	args := pgx.NamedArgs{"id": user_data.ID, "name": user_data.Name, "email": user_data.Email}
	_, err = conn.Exec(context.Background(), query, args)

	if err != nil {
		log.Fatalf("Failed to insert query\n")
		return false
	}
	return true
}
