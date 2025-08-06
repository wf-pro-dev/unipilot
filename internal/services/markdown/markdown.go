package markdown

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// MarkdownService provides functionality to parse markdown to HTML
type MarkdownService struct{}

// NewMarkdownService creates a new instance of MarkdownService
func NewMarkdownService() *MarkdownService {
	return &MarkdownService{}
}

// ParseToHTML converts markdown text to HTML with basic styling
func (m *MarkdownService) ParseToHTML(markdownText string) (string, error) {
	if strings.TrimSpace(markdownText) == "" {
		return "", nil
	}

	// Create markdown parser with extensions
	extensions := parser.CommonExtensions
	parser := parser.NewWithExtensions(extensions)

	// Parse markdown to AST
	doc := parser.Parse([]byte(markdownText))

	// Create HTML renderer with options
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	// Render AST to HTML
	htmlBytes := markdown.Render(doc, renderer)

	// Wrap in styled container
	styledHTML := m.wrapWithStyles(string(htmlBytes))

	return styledHTML, nil
}

// ParseToHTMLWithCustomStyles converts markdown to HTML with custom CSS styles
func (m *MarkdownService) ParseToHTMLWithCustomStyles(markdownText, customCSS string) (string, error) {
	html, err := m.ParseToHTML(markdownText)
	if err != nil {
		return "", err
	}

	if customCSS == "" {
		return html, nil
	}

	// Wrap with custom styles
	styledHTML := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		%s
	</style>
</head>
<body>
	%s
</body>
</html>`, customCSS, html)

	return styledHTML, nil
}

// wrapWithStyles wraps HTML content with default markdown styling
func (m *MarkdownService) wrapWithStyles(htmlContent string) string {
	const defaultStyles = `
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
			line-height: 1.6;
			color: #333;
			max-width: 800px;
			margin: 0 auto;
			padding: 20px;
			background-color: #ffffff;
		}
		h1, h2, h3, h4, h5, h6 {
			color: #2c3e50;
			margin-top: 24px;
			margin-bottom: 16px;
			font-weight: 600;
			line-height: 1.25;
		}
		h1 { font-size: 2em; border-bottom: 1px solid #eaecef; padding-bottom: 0.3em; }
		h2 { font-size: 1.5em; border-bottom: 1px solid #eaecef; padding-bottom: 0.3em; }
		h3 { font-size: 1.25em; }
		h4 { font-size: 1em; }
		h5 { font-size: 0.875em; }
		h6 { font-size: 0.85em; color: #6a737d; }
		p { margin-bottom: 16px; }
		blockquote {
			padding: 0 1em;
			color: #6a737d;
			border-left: 0.25em solid #dfe2e5;
			margin: 0 0 16px 0;
		}
		code {
			background-color: rgba(27, 31, 35, 0.05);
			border-radius: 3px;
			font-size: 85%;
			margin: 0;
			padding: 0.2em 0.4em;
			font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
		}
		pre {
			background-color: #f6f8fa;
			border-radius: 3px;
			font-size: 85%;
			line-height: 1.45;
			overflow: auto;
			padding: 16px;
			margin-bottom: 16px;
		}
		pre code {
			background-color: transparent;
			padding: 0;
		}
		table {
			border-collapse: collapse;
			border-spacing: 0;
			width: 100%;
			margin-bottom: 16px;
		}
		table th, table td {
			border: 1px solid #dfe2e5;
			padding: 6px 13px;
		}
		table th {
			background-color: #f6f8fa;
			font-weight: 600;
		}
		ul, ol {
			padding-left: 2em;
			margin-bottom: 16px;
		}
		li {
			margin-bottom: 0.25em;
		}
		hr {
			height: 0.25em;
			padding: 0;
			margin: 24px 0;
			background-color: #e1e4e8;
			border: 0;
		}
		a {
			color: #0366d6;
			text-decoration: none;
		}
		a:hover {
			text-decoration: underline;
		}
		img {
			max-width: 100%;
			height: auto;
		}
		.highlight {
			background-color: #fff3cd;
			padding: 2px 4px;
			border-radius: 3px;
		}
	`

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		%s
	</style>
</head>
<body>
	%s
</body>
</html>`, defaultStyles, htmlContent)
}

// ParseInline converts inline markdown to HTML (without block elements)
func (m *MarkdownService) ParseInline(markdownText string) (string, error) {
	if strings.TrimSpace(markdownText) == "" {
		return "", nil
	}

	// Create parser with inline-only extensions
	extensions := parser.CommonExtensions
	parser := parser.NewWithExtensions(extensions)

	// Parse markdown
	doc := parser.Parse([]byte(markdownText))

	// Create HTML renderer
	htmlFlags := html.CommonFlags
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	// Render to HTML
	htmlBytes := markdown.Render(doc, renderer)

	return string(htmlBytes), nil
}

// GetDefaultStyles returns the default CSS styles for markdown rendering
func (m *MarkdownService) GetDefaultStyles() string {
	return `
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
			line-height: 1.6;
			color: #333;
			max-width: 800px;
			margin: 0 auto;
			padding: 20px;
			background-color: #ffffff;
		}
		h1, h2, h3, h4, h5, h6 {
			color: #2c3e50;
			margin-top: 24px;
			margin-bottom: 16px;
			font-weight: 600;
			line-height: 1.25;
		}
		h1 { font-size: 2em; border-bottom: 1px solid #eaecef; padding-bottom: 0.3em; }
		h2 { font-size: 1.5em; border-bottom: 1px solid #eaecef; padding-bottom: 0.3em; }
		h3 { font-size: 1.25em; }
		h4 { font-size: 1em; }
		h5 { font-size: 0.875em; }
		h6 { font-size: 0.85em; color: #6a737d; }
		p { margin-bottom: 16px; }
		blockquote {
			padding: 0 1em;
			color: #6a737d;
			border-left: 0.25em solid #dfe2e5;
			margin: 0 0 16px 0;
		}
		code {
			background-color: rgba(27, 31, 35, 0.05);
			border-radius: 3px;
			font-size: 85%;
			margin: 0;
			padding: 0.2em 0.4em;
			font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
		}
		pre {
			background-color: #f6f8fa;
			border-radius: 3px;
			font-size: 85%;
			line-height: 1.45;
			overflow: auto;
			padding: 16px;
			margin-bottom: 16px;
		}
		pre code {
			background-color: transparent;
			padding: 0;
		}
		table {
			border-collapse: collapse;
			border-spacing: 0;
			width: 100%;
			margin-bottom: 16px;
		}
		table th, table td {
			border: 1px solid #dfe2e5;
			padding: 6px 13px;
		}
		table th {
			background-color: #f6f8fa;
			font-weight: 600;
		}
		ul, ol {
			padding-left: 2em;
			margin-bottom: 16px;
		}
		li {
			margin-bottom: 0.25em;
		}
		hr {
			height: 0.25em;
			padding: 0;
			margin: 24px 0;
			background-color: #e1e4e8;
			border: 0;
		}
		a {
			color: #0366d6;
			text-decoration: none;
		}
		a:hover {
			text-decoration: underline;
		}
		img {
			max-width: 100%;
			height: auto;
		}
		.highlight {
			background-color: #fff3cd;
			padding: 2px 4px;
			border-radius: 3px;
		}
	`
}

// ParseWithTemplate converts markdown to HTML using a custom template
func (m *MarkdownService) ParseWithTemplate(markdownText, templateStr string) (string, error) {
	html, err := m.ParseToHTML(markdownText)
	if err != nil {
		return "", err
	}

	// Parse template
	tmpl, err := template.New("markdown").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	data := map[string]interface{}{
		"Content": template.HTML(html),
	}

	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
