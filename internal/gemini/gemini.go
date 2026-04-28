package gemini

import "d_assist/internal/types"

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/genai"
	"os"
	"strings"
)

var syllabusSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"course_title": {
			Type:        genai.TypeString,
			Description: "The full descriptive name of the course, excluding the course code (e.g., 'Introduction to Computer Science' rather than 'CS101 Intro').",
		},
		"course_abbr": {
			Type:        genai.TypeString,
			Description: "The course abbreviation and number. If multiple are present, provide the primary one associated with the main syllabus. Example: 'BIO101'.",
		},
		"assignments": {
			Type:        genai.TypeArray,
			Description: "A comprehensive list of graded deliverables, exams, and key deadlines extracted from the course schedule or syllabus.",
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"title": {
						Type:        genai.TypeString,
						Description: "The specific name of the task (e.g., 'Midterm Exam', 'Problem Set 1'). Do not include the date in the title.",
					},
					"dueDate": {
						Type:        genai.TypeString,
						Description: "The deadline exactly as written. If a specific year isn't mentioned, assume the current academic year. Use 'TBD' if the date is not listed.",
					},
					"weight": {
						Type:        genai.TypeString,
						Description: "The grade contribution (e.g., '20%'). If the syllabus provides points instead of percentages, include the point value (e.g., '50 pts').",
					},
				},
				Required: []string{"title", "dueDate", "weight"},
			},
		},
	},
	Required: []string{"course_title", "course_abbr", "assignments"},
}

func create_gemini_client() (*genai.Client, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: os.Getenv("GOOGLE_AI_API_KEY"),
	})
	if err != nil {
		err = fmt.Errorf("Genai Client creation error %v\n", err)
		fmt.Println(err.Error())

		return nil, err
	}
	return client, nil
}

var model = "gemini-2.5-flash"

func Extract_syllabus(pdf *[]byte) (*da_types.Syllabus, error) {
	gem_client, err := create_gemini_client()
	if err != nil {
		err = fmt.Errorf("Genai Client creation error %v\n", err)
		fmt.Println(err.Error())
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
		err = fmt.Errorf("what happened to the response %v\n", err)
		fmt.Println(err.Error())

		return nil, err
	}
	if len(response.Candidates) == 0 {
		err = fmt.Errorf("%v\n", err)
		fmt.Println(err.Error())

		return nil, err
	}
	validate_pdf := fmt.Sprint(response.Candidates[0].Content.Parts[0])

	// Check for the exact error string (using strings.TrimSpace to remove any accidental newlines)
	if strings.TrimSpace(validate_pdf) == "ERROR_NOT_A_SYLLABUS" {
		err = fmt.Errorf("Validation Failed: The user did not upload a syllabus.")
		fmt.Println(err.Error())

		return nil, err
	}

	json_string := string(response.Candidates[0].Content.Parts[0].Text)

	var syllabus da_types.Syllabus
	json.Unmarshal([]byte(json_string), &syllabus)
	return &syllabus, nil
}
