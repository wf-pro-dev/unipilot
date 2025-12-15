package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"google.golang.org/genai"

	"unipilot/internal/secrets"
)

type GeminiRequest struct {
	Title      string `json:"title"`
	Subject    string `json:"subject"`
	CourseName string `json:"course_name"`
}

type GeminiResponse struct {
	Keywords string
	Content  string
}

func GenerateNote(request *GeminiRequest) (*GeminiResponse, error) {
	ctx := context.Background()

	GEMINI_API_KEY, err := secrets.GetEnvVar("GEMINI_API_KEY")
	if err != nil {
		return nil, err
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  GEMINI_API_KEY,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	config := &genai.GenerateContentConfig{
		MaxOutputTokens:  16384,
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
		"gemini-2.5-flash-lite",
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

func GenerateNoteStream(request *GeminiRequest, writer func(string) error) error {
	ctx := context.Background()

	GEMINI_API_KEY, err := secrets.GetEnvVar("GEMINI_API_KEY")
	if err != nil {
		return err
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  GEMINI_API_KEY,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return err
	}

	config := &genai.GenerateContentConfig{
		MaxOutputTokens: 16384,
	}

	prompt, err := GetPrompt(request)
	if err != nil {
		return errors.New("failed to get prompt")
	}

	stream := client.Models.GenerateContentStream(
		ctx,
		"gemini-2.5-flash-lite",
		genai.Text(prompt),
		config,
	)

	for chunk := range stream {
		if len(chunk.Candidates) > 0 && len(chunk.Candidates[0].Content.Parts) > 0 {
			part := chunk.Candidates[0].Content.Parts[0]
			chunkText := part.Text

			// Send chunk as Markdown text
			if err := writer(chunkText); err != nil {
				return fmt.Errorf("failed to write chunk: %w", err)
			}
		}
	}

	return nil
}

func GetPrompt(request *GeminiRequest) (string, error) {
	prompt := fmt.Sprintf("You are an expert academic tutor and note-taking specialist. Your task is to generate comprehensive, well-structured lecture notes in Markdown format based on the provided information.\n\n"+
		"CONTEXT:\n"+
		"- Course: %s\n"+
		"- Subject: %s\n"+
		"- Lecture Title: %s\n\n"+
		"INSTRUCTIONS:\n"+
		"1. Create a comprehensive lecture summary with a minimum of 3000 words\n"+
		"2. Structure the content in proper Markdown format with clear headings and subheadings\n"+
		"3. Include key concepts, definitions, examples, and important points\n"+
		"4. Use academic language appropriate for the subject level\n"+
		"5. Use schematics and visual representation when you can\n"+
		"6. Organize information logically with bullet points and numbered lists where appropriate\n"+
		"7. Include any relevant formulas, equations, or technical terms\n"+
		"8. Add contextual explanations for complex concepts\n"+
		"9. Ensure the content is educational and informative\n\n"+
		"OUTPUT FORMAT:\n"+
		"Generate the complete lecture notes directly in Markdown format. Do not include any JSON wrapper or additional fields - output only the Markdown content.\n\n"+
		"MARKDOWN SYNTAX EXAMPLES - USE THESE FORMATS:\n\n"+
		"# Heading 1 (Main Title)\n"+
		"## Heading 2 (Major Section)\n"+
		"### Heading 3 (Subsection)\n"+
		"#### Heading 4 (Sub-subsection)\n"+
		"##### Heading 5\n"+
		"###### Heading 6\n\n"+
		"**Bold text** for emphasis\n"+
		"*Italic text* for subtle emphasis\n"+
		"***Bold and italic*** for strong emphasis\n"+
		"`inline code` for technical terms\n"+
		"~~Strikethrough~~ for deprecated concepts\n\n"+
		"Unordered lists:\n"+
		"- First item\n"+
		"- Second item\n"+
		"  - Nested item\n"+
		"  - Another nested item\n"+
		"- Third item\n\n"+
		"Ordered lists:\n"+
		"1. First step\n"+
		"2. Second step\n"+
		"   1. Sub-step\n"+
		"   2. Another sub-step\n"+
		"3. Third step\n\n"+
		"Task lists:\n"+
		"- [ ] Incomplete task\n"+
		"- [x] Completed task\n\n"+
		"Tables:\n"+
		"| Column 1 | Column 2 | Column 3 |\n"+
		"|----------|----------|----------|\n"+
		"| Data 1   | Data 2   | Data 3   |\n"+
		"| Value A  | Value B  | Value C  |\n\n"+
		"Code blocks with syntax highlighting:\n"+
		"```python\n"+
		"def example_function():\n"+
		"    return \"Hello, World!\"\n"+
		"```\n\n"+
		"```javascript\n"+
		"function example() {\n"+
		"    console.log(\"Hello, World!\");\n"+
		"}\n"+
		"```\n\n"+
		"Blockquotes:\n"+
		"> This is an important quote or note.\n"+
		"> It can span multiple lines.\n\n"+
		"Links:\n"+
		"[Link text](https://example.com)\n\n"+
		"Images:\n"+
		"![Alt text](image-url.png)\n\n"+
		"Horizontal rule:\n"+
		"---\n\n"+
		"LaTeX Math (inline): $E = mc^2$ or $\\int_0^\\infty e^{-x^2} dx = \\frac{\\sqrt{\\pi}}{2}$\n\n"+
		"LaTeX Math (block):\n"+
		"$$\n"+
		"\\sum_{i=1}^{n} x_i = x_1 + x_2 + \\cdots + x_n\n"+
		"$$\n\n"+
		"Mermaid Diagrams:\n"+
		"```mermaid\n"+
		"graph TD\n"+
		"    A[Start] --> B{Decision}\n"+
		"    B -->|Yes| C[Action 1]\n"+
		"    B -->|No| D[Action 2]\n"+
		"    C --> E[End]\n"+
		"    D --> E\n"+
		"```\n\n"+
		"```mermaid\n"+
		"sequenceDiagram\n"+
		"    participant A as User\n"+
		"    participant B as System\n"+
		"    A->>B: Request\n"+
		"    B-->>A: Response\n"+
		"```\n\n"+
		"Definition lists (using HTML):\n"+
		"<dl>\n"+
		"<dt>Term 1</dt>\n"+
		"<dd>Definition of term 1</dd>\n"+
		"<dt>Term 2</dt>\n"+
		"<dd>Definition of term 2</dd>\n"+
		"</dl>\n\n"+
		"CONTENT REQUIREMENTS:\n"+
		"- Start with a clear introduction that sets the context\n"+
		"- Include main topics and subtopics with proper headings (use ## for main sections, ### for subsections)\n"+
		"- Provide clear explanations of key concepts using **bold** for important terms\n"+
		"- Include relevant examples or case studies in code blocks or blockquotes\n"+
		"- Add a brief summary or conclusion section\n"+
		"- Use proper academic formatting and citation style where appropriate\n"+
		"- Ensure the content flows logically and is easy to follow\n"+
		"- Include any important definitions or terminology explanations\n"+
		"- Add practical applications or real-world connections where relevant\n"+
		"- Use tables for comparisons, data, or structured information\n"+
		"- Use code blocks for algorithms, pseudocode, or technical implementations\n"+
		"- Use blockquotes for important notes, quotes, or callouts\n\n"+
		"VISUAL REPRESENTATION REQUIREMENTS:\n"+
		"- **Mermaid Diagrams**: Use for processes, workflows, relationships, hierarchies, timelines\n"+
		"  - Example: ```mermaid\\ngraph LR\\nA --> B\\nB --> C\\n```\n"+
		"- **Mathematical Notation**: Use LaTeX format ($inline$ and $$block$$) for all equations\n"+
		"  - Inline: $x = \\frac{-b \\pm \\sqrt{b^2-4ac}}{2a}$\n"+
		"  - Block: $$\\int_a^b f(x)dx = F(b) - F(a)$$\n"+
		"- **Tables**: For comparisons, data, classifications, or structured information\n"+
		"- **Code Blocks**: For algorithms, pseudocode, or technical implementations (specify language)\n"+
		"- **Flowcharts**: For decision processes, step-by-step procedures (use Mermaid)\n"+
		"- **Mind Maps**: For concept relationships and connections (use Mermaid)\n"+
		"- **Sequence Diagrams**: For interactions, communications, or temporal processes (use Mermaid)\n"+
		"- **Entity Relationship Diagrams**: For data structures or system relationships (use Mermaid)\n\n"+
		"QUALITY STANDARDS:\n"+
		"- Academic rigor and accuracy\n"+
		"- Clear, concise writing\n"+
		"- Logical organization with proper heading hierarchy\n"+
		"- Comprehensive coverage of the topic\n"+
		"- Educational value for students\n"+
		"- Professional presentation\n"+
		"- Consistent use of Markdown formatting throughout\n"+
		"- Proper spacing between sections (blank lines between paragraphs)\n"+
		"- Use of formatting to enhance readability (bold for key terms, lists for multiple points)\n\n"+
		"Please generate high-quality, comprehensive notes in Markdown format that would be valuable for students studying this subject. Use the examples above to ensure proper formatting and structure.",
		request.CourseName, request.Subject, request.Title)

	return prompt, nil
}
