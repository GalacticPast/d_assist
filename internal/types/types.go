package da_types

const (
	SUPABASE_PUBLIC_CLIENT = iota
	SUPABASE_ADMIN_CLIENT  // 1
)

type Supabase_client_type int

type User_info struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type User_creds struct {
	First_Name string `json:"user_first_name"`
	Last_Name  string `json:"user_last_name"`
	Email      string `json:"user_email"`
	Password   string `json:"user_password"`
}
