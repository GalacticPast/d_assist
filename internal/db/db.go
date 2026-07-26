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
	"strconv"
	"strings"
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

func Get_courses(user_id string) (*[]da_types.Course, error) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
		runtime.Breakpoint()
	}
	defer conn.Close(context.Background())

	query := "SELECT name, code, color FROM courses WHERE user_id = (@user_id)"

	args := pgx.NamedArgs{"user_id": user_id}

	rows, err := conn.Query(context.Background(), query, args)
	if err != nil {
		return nil, err
	}

	courses, err := pgx.CollectRows(rows, pgx.RowToStructByName[da_types.Course])
	if err != nil {
		return nil, fmt.Errorf("failed to collect rows: %w", err)
	}

	return &courses, nil
}

func Get_signed_upload_url(button_id string, file_path string) string {
	supabase_client, err := Create_supabase_client(da_types.SUPABASE_ADMIN_CLIENT)
	if err != nil {
		log.Printf("Failed to initalize the client: %v\n", err)
		return ""
	}
	bucket_name := "slides_pdf"
	if button_id == "Syllabus-upload-btn" {
		bucket_name = "syllabus_pdf"
	}
	resp, err := supabase_client.Storage.CreateSignedUploadUrl(bucket_name, file_path)
	if err != nil {
		log.Printf("Failed to get the upload url: %v\n", err)
		return ""
	}
	// @warn: do I have to do some additional signing??
	return resp.Url
}

func Get_pdf_from_bucket(button_id string, file_name string) ([]byte, error) {
	supabase_client, err := Create_supabase_client(da_types.SUPABASE_ADMIN_CLIENT)
	if err != nil {
		log.Printf("Failed to initalize the client: %v\n", err)
		return nil, err
	}
	bucket_name := "slides_pdf"
	if button_id == "Syllabus-upload-btn" {
		bucket_name = "syllabus_pdf"
	}
	pdf_bytes, err := supabase_client.Storage.DownloadFile(bucket_name, file_name)
	if err != nil {
		fmt.Println(err.Error())

		err = fmt.Errorf("Supabase error: %v\n", err)
		return nil, err
	}
	return pdf_bytes, nil
}

func Insert_new_deck(user_id string, course_badge string, deck *da_types.Deck) error {
	supabase_client, err := Create_supabase_client(da_types.SUPABASE_ADMIN_CLIENT)
	if err != nil {
		log.Printf("Failed to initalize the client: %v\n", err)
		return err
	}

	var deck_result []map[string]interface{}

	_, err = supabase_client.From("decks").Insert(map[string]interface{}{
		"course_id": user_id,
		"title":     deck.Title,
	}, false, "", "", "").ExecuteTo(&deck_result)

	if err != nil {
		return fmt.Errorf("failed to insert course: %w", err)
	}

	if len(deck_result) == 0 {
		return fmt.Errorf("course inserted but no ID was returned")
	}
	deck_id := deck_result[0]["id"].(string)

	var batch_cards []map[string]interface{}

	for _, a := range deck.Cards {
		batch_cards = append(batch_cards, map[string]interface{}{
			"deck_id":     deck_id, // Links to the parent course
			"question":    a.Question,
			"options":     a.Options,
			"awnser":      a.Correct_option,
			"explanation": a.Explanation,
		})
	}

	_, _, err = supabase_client.From("cards").Insert(batch_cards, false, "", "", "").Execute()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	log.Printf("Successfully inserted course and %d assignments.", len(batch_cards))
	return nil
}

func Insert_new_syllabus(user_id string, syllabus *da_types.Syllabus) error {
	supabase_client, err := Create_supabase_client(da_types.SUPABASE_ADMIN_CLIENT)
	if err != nil {
		log.Printf("Failed to initalize the client: %v\n", err)
		return err
	}

	var courseResult []map[string]interface{}

	_, err = supabase_client.From("courses").Insert(map[string]interface{}{
		"user_id": user_id,
		"name":    syllabus.Course_title,
		"code":    syllabus.Course_abbr,
	}, false, "", "", "").ExecuteTo(&courseResult)

	if err != nil {
		return fmt.Errorf("failed to insert course: %w", err)
	}

	// Extract the UUID of the newly created course
	if len(courseResult) == 0 {
		return fmt.Errorf("course inserted but no ID was returned")
	}
	courseID := courseResult[0]["id"].(string)

	// 3. Prepare the Assignments for Batch Insertion
	var batchAssignments []map[string]interface{}

	// Get the current year to append to dates (since syllabi rarely include the year)
	currentYear := time.Now().Year()

	for _, a := range syllabus.Assignments {
		// --- CLEANUP: Weight ---
		// Remove '%' and convert to float for the NUMERIC(5,2) column
		weightClean := strings.TrimSpace(strings.TrimSuffix(a.Weight, "%"))
		weightNum, _ := strconv.ParseFloat(weightClean, 64)
		// Note: If ParseFloat fails (e.g., if weight is "N/A"), it defaults to 0.00, which is safe for the DB.

		// --- CLEANUP: Date ---
		// LLMs usually return dates like "Oct 20". We append the year and parse it.
		dateStringWithYear := fmt.Sprintf("%s %d", a.Due_date, currentYear)
		parsedDate, parseErr := time.Parse("Jan 02 2006", dateStringWithYear)

		if parseErr != nil {
			// Fallback: If the LLM returned garbage like "TBD", default the date to 1 month from now
			// so the DB doesn't reject the insert (since due_date is NOT NULL)
			parsedDate = time.Now().AddDate(0, 1, 0)
		}

		// --- APPEND TO BATCH ---
		batchAssignments = append(batchAssignments, map[string]interface{}{
			"course_id": courseID, // Links to the parent course
			"title":     a.Title,
			"due_date":  parsedDate,
			"weight":    weightNum,
			// difficulty defaults to 3 via SQL, so we can omit it
		})
	}

	// 4. Execute the Batch Insert
	// Passing the slice of maps automatically triggers a multi-row INSERT in Supabase
	supabase_client.From("assignments").Insert(batchAssignments, false, "", "", "").Execute()

	log.Printf("Successfully inserted course and %d assignments.", len(batchAssignments))
	return nil
}
