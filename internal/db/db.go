package db

import "d_assist/internal/types"

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/supabase-community/gotrue-go/types"
	"github.com/supabase-community/supabase-go"
	"log"
	"os"
	"runtime"
	"time"
)

// @todo: hmmm how to make this one time only, cause for the duration of this app this should only create 2 clients.
// we should just be reusing it right?
func Create_supabase_client(client_type da_types.Supabase_client_type) (*supabase.Client, error) {

	supabase_pub_url := os.Getenv("NEXT_PUBLIC_SUPABASE_URL")

	var client *supabase.Client = nil
	var err error = nil

	switch client_type {
	case da_types.SUPABASE_PUBLIC_CLIENT:
		supabase_pub_anon_key := os.Getenv("NEXT_PUBLIC_SUPABASE_ANON_KEY")
		client, err = supabase.NewClient(supabase_pub_url, supabase_pub_anon_key, nil)
	case da_types.SUPABASE_ADMIN_CLIENT:
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
	supabase_client, err := Create_supabase_client(da_types.SUPABASE_PUBLIC_CLIENT)
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
	sb_client, err := Create_supabase_client(da_types.SUPABASE_PUBLIC_CLIENT)
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

func Check_if_user_exists(user_data *da_types.User_info) bool {
	// @todo: have a connection pool
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer conn.Close(context.Background())

	query := "SELECT EXISTS(SELECT 1 FROM profiles WHERE id = $1)"

	var exists bool
	err = conn.QueryRow(context.Background(), query, user_data.ID).Scan(&exists)

	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
		runtime.Breakpoint()
	}
	if exists {
		return true
	}
	return false
}

func Get_JWT_Token(user_info *da_types.User_info) (string, error) {

	claims := jwt.MapClaims{
		"role":  "authenticated", // Tells Supabase this is a logged-in user
		"sub":   user_info.ID,    // The unique ID. Supabase auth.uid() will equal this value!
		"email": user_info.Email,
		"iss":   "https://wtpfmvqjwzkwtsvswtmm.supabase.co/auth/v1",
		"aud":   "authenticated",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = "1ab49d83-51e2-4967-86db-6f2da1309f90"
	// Sign the token using your Supabase JWT Secret
	signedToken, err := token.SignedString([]byte(os.Getenv("SUPABASE_JWT_KEY")))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func Create_user(user_data *da_types.User_info) bool {
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

func Get_number_of_courses(user_id string) int {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
		runtime.Breakpoint()
	}
	defer conn.Close(context.Background())

	query := "SELECT COUNT(*) FROM courses WHERE user_id = (@user_id)"

	args := pgx.NamedArgs{"user_id": user_id}

	var count int
	err = conn.QueryRow(context.Background(), query, args).Scan(&count)

	if err != nil {
		log.Fatalf("Failed to count courses: %v", err)
		return 0
	}
	return count
}

var bucket_name = "syllabus_pdf"

func Get_signed_upload_url(file_path string) string {
	supabase_client, err := Create_supabase_client(da_types.SUPABASE_ADMIN_CLIENT)
	if err != nil {
		log.Printf("Failed to initalize the client: %v\n", err)
		return ""
	}
	resp, err := supabase_client.Storage.CreateSignedUploadUrl(bucket_name, file_path)
	if err != nil {
		log.Printf("Failed to get the upload url: %v\n", err)
		return ""
	}
	// @warn: do I have to do some additional signing??
	return resp.Url
}

func Get_pdf_from_bucket(file_name string) ([]byte, error) {
	supabase_client, err := Create_supabase_client(da_types.SUPABASE_ADMIN_CLIENT)
	if err != nil {
		log.Printf("Failed to initalize the client: %v\n", err)
		return nil, err
	}
	pdf_bytes, err := supabase_client.Storage.DownloadFile("syllabus_pdf", file_name)
	if err != nil {
		fmt.Errorf("Supabase error: %v\n", err)
		return nil, err
	}
	return pdf_bytes, nil
}
