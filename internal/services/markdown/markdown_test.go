package markdown

import (
	"fmt"
	"testing"
)

func TestMarkdownService_ParseToHTML(t *testing.T) {
	service := NewMarkdownService()

	// Test basic markdown parsing
	markdownText := `# Hello World

This is a **bold** text and *italic* text.

## Code Example

` + "```" + `go
func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

## List Example

- Item 1
- Item 2
  - Subitem 2.1
  - Subitem 2.2

## Table Example

| Name | Age | City |
|------|-----|------|
| John | 25  | NYC  |
| Jane | 30  | LA   |

> This is a blockquote with some important information.

[Link to Google](https://www.google.com)
`

	html, err := service.ParseToHTML(markdownText)
	if err != nil {
		t.Fatalf("Failed to parse markdown: %v", err)
	}

	if html == "" {
		t.Error("Expected HTML output, got empty string")
	}

	fmt.Println("Generated HTML:")
	fmt.Println(html)
}

func TestMarkdownService_ParseInline(t *testing.T) {
	service := NewMarkdownService()

	// Test inline markdown parsing
	inlineText := "This is **bold** and *italic* text with " + "`" + "code" + "`" + " and [link](https://example.com)"

	html, err := service.ParseInline(inlineText)
	if err != nil {
		t.Fatalf("Failed to parse inline markdown: %v", err)
	}

	if html == "" {
		t.Error("Expected HTML output, got empty string")
	}

	fmt.Println("Inline HTML:")
	fmt.Println(html)
}

func TestMarkdownService_ParseWithCustomStyles(t *testing.T) {
	service := NewMarkdownService()

	markdownText := `# Custom Styled Content

This content will have custom CSS styling applied.

## Features

- **Bold text**
- *Italic text*
- ` + "`" + `Inline code` + "`" + `

` + "```" + `javascript
console.log("Hello from custom styles!");
` + "```" + `
`

	customCSS := `
		body {
			font-family: 'Courier New', monospace;
			background-color: #1a1a1a;
			color: #00ff00;
			padding: 20px;
		}
		h1, h2 {
			color: #ff6600;
			border-bottom: 2px solid #ff6600;
		}
		code {
			background-color: #333;
			color: #00ff00;
			border: 1px solid #666;
		}
		pre {
			background-color: #222;
			border: 1px solid #444;
		}
	`

	html, err := service.ParseToHTMLWithCustomStyles(markdownText, customCSS)
	if err != nil {
		t.Fatalf("Failed to parse with custom styles: %v", err)
	}

	if html == "" {
		t.Error("Expected HTML output, got empty string")
	}

	fmt.Println("Custom Styled HTML:")
	fmt.Println(html)
}
