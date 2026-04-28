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

type Task struct {
	Title          string
	Due_date       string
	Progress       string // its going go be n%
	Progress_color string // green for 100% and dark for everything else
}

type Course struct {
	Title   string `db:"name"`
	Color   string `db:"color"`
	Badge   string `db:"code"`
	Tasks   []Task `db:"-"`
	Quizzes []Quiz `db:"-"`
}

type Syllabus struct {
	Course_title string       `json:"course_title"`
	Course_abbr  string       `json:"course_abbr"`
	Assignments  []Assignment `json:"assignments"`
}
type Quiz struct {
	Quiz_id string `db:"quiz_id"`
	Title   string `db:"quiz_title"`
}

type Assignment struct {
	Title    string `json:"title"`
	Due_date string `json:"dueDate"`
	Weight   string `json:"weight"`
}
