package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	
	"github.com/spf13/viper"

	"google.golang.org/genai"
)

type GeminiRequest struct {
	Title      string `json:"title"`
	Subject    string `json:"subject"`
	CourseName string `json:"course_name"`
}

type GeminiResponse struct {
	Keywords	string
	Content		string
}

func GenerateNote(request *GeminiRequest) (*GeminiResponse, error) {
	ctx := context.Background()
	

	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	GEMINI_API_KEY := viper.GetString("GEMINI_API_KEY")


	client, err := genai.NewClient(ctx, &genai.ClientConfig{
        	APIKey:  GEMINI_API_KEY,
        	Backend: genai.BackendGeminiAPI,
    	})
	if err != nil {
		log.Fatal(err)
	}

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"keywords": {Type: genai.TypeString},
				"content":  {Type: genai.TypeString},
			},
			PropertyOrdering: []string{"keywords", "content"},
		},
	}

	prompt, err := GetPrompt(request)
	if err != nil {
		return nil, errors.New("failed to get prompt")
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-1.5-flash",
		genai.Text(prompt),
		config,
	)
	if err != nil {
		log.Fatal(err)
	}


	var response *GeminiResponse
	err = json.Unmarshal([]byte(result.Text()), &response)
	if err != nil {
		return nil, errors.New("failed to unmarshal response")
	}



	return response, nil
}

func GetPrompt(request *GeminiRequest) (string, error) {
	prompt := fmt.Sprintf(`You are an expert academic tutor and note-taking specialist. Your task is to generate comprehensive, well-structured lecture notes based on the provided information.

CONTEXT:
- Course: %s
- Subject: %s
- Lecture Title: %s

INSTRUCTIONS:
1. Generate exactly 5 relevant keywords that capture the main concepts of this lecture
2. Create a comprehensive lecture summary with a minimun of 2000 words
3. Structure the content in proper Markdown format with clear headings and subheadings
4. Include key concepts, definitions, examples, and important points
5. Use academic language appropriate for the subject level
6. Use schematics and visual representation when you can
7. Organize information logically with bullet points and numbered lists where appropriate
8. Include any relevant formulas, equations, or technical terms
9. Add contextual explanations for complex concepts
10. Ensure the content is educational and informative

OUTPUT FORMAT:
Return a JSON object with exactly two fields:
- 'keywords': A comma-separated string of exactly 5 keywords
- 'content': The complete lecture notes in Markdown format

CONTENT REQUIREMENTS:
- Start with a clear introduction that sets the context
- Include main topics and subtopics with proper headings
- Provide clear explanations of key concepts
- Include relevant examples or case studies
- Add a brief summary or conclusion
- Use proper academic formatting and citation style where appropriate
- Ensure the content flows logically and is easy to follow
- Include any important definitions or terminology explanations
- Add practical applications or real-world connections where relevant

VISUAL REPRESENTATION REQUIREMENTS:
- **Mermaid Diagrams**: Use for processes, workflows, relationships, hierarchies, timelines
- **Mathematical Notation**: Use LaTeX format ($inline$ and $$block$$) for all equations
- **Tables**: For comparisons, data, classifications, or structured information
- **Code Blocks**: For algorithms, pseudocode, or technical implementations
- **Flowcharts**: For decision processes, step-by-step procedures
- **Mind Maps**: For concept relationships and connections
- **Sequence Diagrams**: For interactions, communications, or temporal processes
- **Entity Relationship Diagrams**: For data structures or system relationships

QUALITY STANDARDS:
- Academic rigor and accuracy
- Clear, concise writing
- Logical organization
- Comprehensive coverage of the topic
- Educational value for students
- Professional presentation

Please generate high-quality, comprehensive notes that would be valuable for students studying this subject.`,
		request.CourseName, request.Subject, request.Title)

	return prompt, nil
}
