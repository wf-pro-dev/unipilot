package document

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unipilot/internal/secrets"

	"github.com/google/uuid"

	"unipilot/internal/errors"

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

func GetFileText(filename, ext string) (string, error) {
	if ext == ".docx" {
		return GetFileTextForDocx(filename)
	}
	if ext == ".pdf" {
		return GetFileTextForPdf(filename)
	}

	return "", errors.Wrap(fmt.Errorf("file type not supported"), errors.FSFileTypeNotSupported, "File type not supported")
}

func GetFileChunks(text string) []string {

	// Remove space and new lines from the text
	text = strings.ReplaceAll(text, "\n", "")

	var chunks []string
	runes := []rune(text)
	chunkSize := 1024                             // 1024 characters per chunk
	overlapSize := int(float64(chunkSize) * 0.15) // 15% overlap

	for i := 0; i < len(runes); i += chunkSize - overlapSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))

		if end == len(runes) {
			break
		}

	}
	return chunks
}

func GetQdrantVectors(document *LocalDocument) ([]*qdrant.PointStruct, error) {
	fmt.Println("GetQdrantVectors")
	var vectors []*qdrant.PointStruct

	//Get file text
	text, err := GetFileText(document.FileName, filepath.Ext(document.FileName))
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
