package gemini

import "d_assist/internal/types"

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/genai"
	"os"
)

var slides_schema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"title": {
			Type:        genai.TypeString,
			Description: "The full descriptive name of the slides.",
		},
		"status": {
			Type:        genai.TypeString,
			Description: "Returns 'SUCCESS' if the input is a valid class slide. Returns 'ERROR_NOT_A_CLASS_SLIDE' if it is not.",
		},
		"cards": {
			Type:        genai.TypeArray,
			Description: "A comprehensive list of quiz questions extracted from the slides. Empty if status is not SUCCESS.",
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"question": {
						Type:        genai.TypeString,
						Description: "The text of the quiz question.",
					},
					"options": {
						Type:        genai.TypeArray,
						Description: "A list of four possible options/answers.",
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
					},
					"correct_option": {
						Type:        genai.TypeInteger,
						Description: "The 0-based index of the correct option in the options array.",
					},
					"explanation": {
						Type:        genai.TypeString,
						Description: "why the correct answer is right",
					},
				},
				Required: []string{"question", "options", "correct_option", "explanation"},
			},
		},
	},
	Required: []string{"title", "status", "cards"},
}

var syllabus_schema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"status": {
			Type:        genai.TypeString,
			Description: "Returns 'SUCCESS' if the input is a syllabus. Returns 'ERROR_NOT_A_SYLLABUS' if it is not.",
		},
		"course_title": {
			Type:        genai.TypeString,
			Description: "The full descriptive name of the course.",
		},
		"course_abbr": {
			Type:        genai.TypeString,
			Description: "The course abbreviation and number.",
		},
		"assignments": {
			Type:        genai.TypeArray,
			Description: "A comprehensive list of graded deliverables.",
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"title": {
						Type:        genai.TypeString,
						Description: "The specific name of the task.",
					},
					"dueDate": {
						Type:        genai.TypeString,
						Description: "The deadline exactly as written.",
					},
					"weight": {
						Type:        genai.TypeString,
						Description: "The grade contribution.",
					},
				},
				Required: []string{"title", "dueDate", "weight"},
			},
		},
	},
	Required: []string{"status", "course_title", "course_abbr", "assignments"},
}

var model = "gemini-2.5-flash"

func create_gemini_client() (*genai.Client, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: os.Getenv("GOOGLE_AI_API_KEY"),
	})
	if err != nil {
		return nil, fmt.Errorf("Genai Client creation error: %w", err)
	}
	return client, nil
}

func Extract_slides(course_badge string, pdf *[]byte) (*da_types.Deck, error) {
	gem_client, err := create_gemini_client()
	if err != nil {
		return nil, err
	}
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   slides_schema,
	}

	promt := fmt.Sprintf(`EVALUATE INPUT: If the document is NOT a class slide/lecture, set "status" to "ERROR_NOT_A_CLASS_SLIDE" and return an empty array.
	If it IS a class slide, set "status" to "SUCCESS". If you think the extracted content doesnt coreespond to course badge: %s,
set "ERROR_SLIDE_BADGE_MISMATCH".
You are an expert instructional designer. Analyze the slides and generate a comprehensive multiple-choice quiz.
1. Generate clear, well-written questions.
2. Provide exactly four (4) plausible options per question.
3. Only one definitively correct answer.
4. Indicate the correct answer using a 0-based index.
5. Ensure distractors are reasonable.`, course_badge)

	promt_part := genai.NewPartFromText(promt)
	pdf_part := genai.NewPartFromBytes(*pdf, "application/pdf")
	contents := []*genai.Content{{Parts: []*genai.Part{promt_part, pdf_part}}}

	response, err := gem_client.Models.GenerateContent(context.Background(), model, contents, config)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from model")
	}

	json_string := response.Candidates[0].Content.Parts[0].Text

	var result da_types.Deck
	if err := json.Unmarshal([]byte(json_string), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Safely check the status field from the JSON
	if result.Status == "ERROR_NOT_A_CLASS_SLIDE" {
		return nil, fmt.Errorf("Validation Failed: The user did not upload a valid class slide")
	} else if result.Status == "ERROR_COURSE_BADGE_MISMATCH" {
		return nil, fmt.Errorf("Validation Failed: The user did not upload a valid class slide corresponding to the course badge")
	}

	return &result, nil
}

func Extract_syllabus(pdf *[]byte) (*da_types.Syllabus, error) {
	gem_client, err := create_gemini_client()
	if err != nil {
		return nil, err
	}
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   syllabus_schema,
	}

	promt := `EVALUATE INPUT: If the document is clearly NOT a course syllabus, set the "status" field to "ERROR_NOT_A_SYLLABUS" and leave the rest blank.
If the document IS a syllabus, set "status" to "SUCCESS". 
You are an elite academic assistant. Analyze the syllabus carefully.
Extract all major assignments, exams, and deliverables. Determine their due dates and weight.
If a due date is implied, estimate it or write 'TBA'.`

	promt_part := genai.NewPartFromText(promt)
	pdf_part := genai.NewPartFromBytes(*pdf, "application/pdf")
	contents := []*genai.Content{{Parts: []*genai.Part{promt_part, pdf_part}}}

	response, err := gem_client.Models.GenerateContent(context.Background(), model, contents, config)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from model")
	}

	json_string := response.Candidates[0].Content.Parts[0].Text

	var result da_types.Syllabus
	if err := json.Unmarshal([]byte(json_string), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Safely check the status field from the JSON
	if result.Status == "ERROR_NOT_A_SYLLABUS" {
		return nil, fmt.Errorf("Validation Failed: The user did not upload a syllabus")
	}

	return &result, nil
}
