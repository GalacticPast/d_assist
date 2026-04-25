package gemini

import (
	"context"
	"fmt"
	"google.golang.org/genai"
	"os"
)

func create_gemini_client() *genai.Client {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: os.Getenv("GOOGLE_AI_API_KEY"),
	})
	if err != nil {
		fmt.Printf("Genai Client creation error %v\n", err)
		return nil
	}
	return client
}

func Get_courses(pdf []byte) {

}
