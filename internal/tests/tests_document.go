package main

import (
	"log"
	"unipilot/internal/models/document"
)

func main() {

	doc := document.Document{
		AssignmentID: uint(10),
		FileName:     "test.docx",
		FileType:     "docx",
	}
	doc.ID = uint(1)

	_, err := document.GetQdrantVectors(&doc)
	if err != nil {
		log.Printf("err: %s", err.Error())
	}
}
