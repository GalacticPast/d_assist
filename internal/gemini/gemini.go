package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/genai"
	"os"
)

type Syllabus_schema struct {
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
		"important_dates": {
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
	Required: []string{"important_dates", "course_title"},
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

func Extract_courses(pdf *[]byte) {
	gem_client, err := create_gemini_client()
	if err != nil {
		fmt.Errorf("Genai Client creation error %v\n", err)
		return
	}
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   syllabusSchema,
	}
	promt := "You are an elite academic assistant. Analyze this course syllabus carefully. Extract all major assignments, exams, and deliverables. Determine their due dates and what percentage of the final grade they are worth. If a due date is not explicitly stated but implied (e.g. week 3), estimate it based on standard semester schedules or just write 'TBA'"
	promt_part := genai.NewPartFromText(promt)
	pdf_part := genai.NewPartFromBytes(*pdf, "application/pdf")
	contents := []*genai.Content{
		{Parts: []*genai.Part{promt_part, pdf_part}},
	}

	result, err := gem_client.Models.GenerateContent(context.Background(), model, contents, config)
	if err != nil {
		fmt.Errorf("what happened to the response %v\n", err)
		return
	}
	count := len(result.Candidates)
	if count == 0 {
		fmt.Errorf("What no errors but the response.Candiates count is 0??")
		return
	}
	//@fix: make this robust
	json_data, err := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(json_data))
}
