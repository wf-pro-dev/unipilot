package fileops

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"
	"unipilot/internal/secrets"

	"code.sajari.com/docconv"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/textsplitter"

	"unipilot/internal/errors"
	"unipilot/internal/models"

	"github.com/qdrant/go-client/qdrant"
	"google.golang.org/genai"
)

func GetFileTextForDocx(filename string) (string, error) {
	// Open the file for reading
	file, err := os.Open(filename)
	if err != nil {
		return "", errors.Wrap(err, errors.FSOpenFailed, "Error opening file")
	}
	defer file.Close()

	// Create command and set stdin to the file
	cmd := exec.Command("python3", "/app/scripts/python/extract_docx.py", filename)
	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, errors.SysExecFailed, "Error executing command")
	}
	return string(output), nil
}

func GetFileTextForPPTX(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", errors.Wrap(err, errors.FSOpenFailed, "Error opening file")
	}
	defer file.Close()

	// Use docconv.ConvertPptx to extract text
	text, _, err := docconv.ConvertPptx(file)
	if err != nil {
		return "", errors.Wrap(err, errors.FSOpenFailed, "Error converting PPTX")
	}

	return text, nil
}

func GetFileTextForPdf(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", errors.Wrap(err, errors.FSOpenFailed, "Error opening file")
	}
	defer file.Close()

	// Create command and set stdin to the file
	cmd := exec.Command("python3", "/app/scripts/python/extract_pdf.py", filename)
	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, errors.SysExecFailed, "Error executing command")
	}
	return string(output), nil
}

func GetFileTextForTxt(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", errors.Wrap(err, errors.FSOpenFailed, "Error opening file")
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		return "", errors.Wrap(err, errors.FSFileFailed, "Error reading file")
	}
	return string(content), nil
}

func GetFileText(filename, ext string) (string, error) {

	switch ext {
	case ".txt", ".md", ".go", ".py", ".js", ".json", ".yaml", ".yml", ".html", ".css", ".scss", ".less", ".sass", ".php", ".sql":
		return GetFileTextForTxt(filename)
	case ".docx", ".doc":
		return GetFileTextForDocx(filename)
	case ".pptx", ".ppt":
		return GetFileTextForPPTX(filename)
	case ".pdf":
		return GetFileTextForPdf(filename)
	default:
		return "", errors.Wrap(fmt.Errorf("file type %s not supported", ext), errors.FSFileTypeNotSupported, "File type not supported")
	}
}

func GetFileChunks(text string) []string {
	chunkSize := 1024                             // 1024 characters per chunk
	overlapSize := int(float64(chunkSize) * 0.15) // 15% overlap (approximately 2-3 sentences)

	// Normalize whitespace: collapse multiple spaces/newlines but preserve paragraph breaks
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{}
	}

	// Normalize multiple newlines to double newlines (paragraph breaks)
	// This helps preserve semantic structure from PDFs
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	// Collapse multiple newlines to double newlines
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}

	// Create a RecursiveCharacter splitter with sentence-aware separators
	// Separators are tried in order: paragraph breaks, sentence endings, newlines, spaces, characters
	// Use NewRecursiveCharacter with options to ensure all fields (including LenFunc) are properly initialized
	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(chunkSize),
		textsplitter.WithChunkOverlap(overlapSize),
		textsplitter.WithSeparators([]string{
			"\n\n", // Paragraph breaks (highest priority)
			". ",   // Period followed by space
			"! ",   // Exclamation followed by space
			"? ",   // Question mark followed by space
			".\n",  // Period followed by newline
			"!\n",  // Exclamation followed by newline
			"?\n",  // Question mark followed by newline
			"\n",   // Single newlines
			" ",    // Spaces
			"",     // Character-by-character fallback
		}),
		textsplitter.WithLenFunc(utf8.RuneCountInString),
	)

	// Split the text into chunks
	chunks, err := splitter.SplitText(text)
	if err != nil {
		// Fallback to simple splitting if there's an error
		return []string{text}
	}

	// Filter out empty chunks and trim whitespace
	var filteredChunks []string
	for _, chunk := range chunks {
		chunk = strings.TrimSpace(chunk)
		if chunk != "" {
			filteredChunks = append(filteredChunks, chunk)
		}
	}

	return filteredChunks
}

func GetQdrantVectors(document *models.Document) ([]*qdrant.PointStruct, error) {
	var vectors []*qdrant.PointStruct

	//Get file text
	fsFileName := filepath.Join("/app/uploads/", document.StorageKey)
	text, err := GetFileText(fsFileName, filepath.Ext(document.FileName))
	if err != nil {
		return nil, errors.Wrap(err, errors.QdrantTextError, "Error getting file text")
	}

	chunks := GetFileChunks(string(text))

	embeddings, err := GetTextEmbedding(chunks)
	if err != nil {
		return nil, errors.Wrap(err, errors.QdrantEmbeddingError, "Error getting text embedding")
	}

	for i, embedding := range embeddings {
		newVector := qdrant.PointStruct{
			Id:      qdrant.NewIDUUID(uuid.New().String()),
			Vectors: qdrant.NewVectors(embedding.Values...),
			Payload: qdrant.NewValueMap(map[string]any{
				"user_id":       document.UserID,
				"document_id":   document.ID,
				"assignment_id": document.AssignmentID,
				"chunk_id":      i,
				"chunk_text":    chunks[i],
			}),
		}
		vectors = append(vectors, &newVector)
	}

	return vectors, nil
}

func GetTextEmbedding(chunks []string) ([]*genai.ContentEmbedding, error) {
	fmt.Println("GetTextEmbedding")
	ctx := context.Background()

	GEMINI_API_KEY, err := secrets.GetEnvVar("GEMINI_API_KEY")
	if err != nil {
		return nil, errors.Wrap(err, errors.ConfigEnvVarNotFound, "Error getting Gemini API key")
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: GEMINI_API_KEY,
	})
	if err != nil {
		return nil, errors.Wrap(err, errors.GeminiFailed, "Error creating Gemini client")
	}

	outputDimensionality := int32(768)
	var contents []*genai.Content
	for _, text := range chunks {
		contents = append(contents, genai.NewContentFromText(text, genai.RoleUser))
	}

	result, err := client.Models.EmbedContent(ctx,
		"gemini-embedding-001",
		contents,
		&genai.EmbedContentConfig{OutputDimensionality: &outputDimensionality, TaskType: "RETRIEVAL_DOCUMENT"},
	)
	if err != nil {
		return nil, errors.Wrap(err, errors.GeminiFailed, "Error embedding content")
	}

	return result.Embeddings, nil
}
