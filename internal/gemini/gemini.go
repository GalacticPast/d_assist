package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/genai"
	"os"
	"strings"
)

type Syllabus struct {
	Course_title string       `json:"course_title"`
	Assignments  []Assignment `json:"assignments"`
}
type Assignment struct {
	Title    string `json:"title"`
	Due_date string `json:"dueDate"`
	Weight   string `json:"weight"`
}

var syllabusSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"course_title": {
			Type:        genai.TypeString,
			Description: "The course name extracted from the pdf.",
		},
		"assignments": {
			Type:        genai.TypeArray,
			Description: "A list of all critical assignments, midterms, finals, and readings.",
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"title": {
						Type:        genai.TypeString,
						Description: "The name of the assignment, exam, or deliverable",
					},
					"dueDate": {
						Type:        genai.TypeString,
						Description: "The exact due date, e.g., 'Oct 20' or 'Every Friday'",
					},
					"weight": {
						Type:        genai.TypeString,
						Description: "The percentage of the total grade, e.g., '15%'. Use 'N/A' if unknown.",
					},
				},
				Required: []string{"title", "dueDate", "weight"},
			},
		},
	},
	Required: []string{"course_title", "important_dates"},
}

func create_gemini_client() (*genai.Client, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: os.Getenv("GOOGLE_AI_API_KEY"),
	})
	if err != nil {
		fmt.Errorf("Genai Client creation error %v\n", err)
		return nil, err
	}
	return client, nil
}

var model = "gemini-2.5-flash"

func Extract_courses(pdf *[]byte) (*Syllabus, error) {
	gem_client, err := create_gemini_client()
	if err != nil {
		fmt.Errorf("Genai Client creation error %v\n", err)
		return nil, err
	}
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   syllabusSchema,
	}
	promt := `First, examine the provided document. If the document is clearly NOT a course syllabus, immediately stop and return exactly "ERROR_NOT_A_SYLLABUS" 
	and nothing else. If the document IS a syllabus, you are an elite academic assistant. Analyze the course syllabus carefully.
	Extract all major assignments, exams, and deliverables. Determine their due dates and what percentage of the final grade they are worth.
	If a due date is not explicitly stated but implied (e.g. week 3), estimate it based on standard semester schedules or just write 'TBA'.`

	promt_part := genai.NewPartFromText(promt)
	pdf_part := genai.NewPartFromBytes(*pdf, "application/pdf")
	contents := []*genai.Content{
		{Parts: []*genai.Part{promt_part, pdf_part}},
	}

	response, err := gem_client.Models.GenerateContent(context.Background(), model, contents, config)
	if err != nil {
		fmt.Errorf("what happened to the response %v\n", err)
		return nil, err
	}
	if len(response.Candidates) == 0 {
		fmt.Errorf("%v\n", err)
		return nil, err
	}
	validate_pdf := fmt.Sprint(response.Candidates[0].Content.Parts[0])

	// Check for the exact error string (using strings.TrimSpace to remove any accidental newlines)
	if strings.TrimSpace(validate_pdf) == "ERROR_NOT_A_SYLLABUS" {
		fmt.Errorf("Validation Failed: The user did not upload a syllabus.")
		return nil, err
	}

	json_string := string(response.Candidates[0].Content.Parts[0].Text)

	var syllabus Syllabus
	json.Unmarshal([]byte(json_string), &syllabus)
	return &syllabus, nil
}
