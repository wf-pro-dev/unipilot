package main

import (
	"log"
	"unipilot/internal/models"
	server "unipilot/internal/server/api"
)

func main() {

	localDoc := models.LocalDocument{
		BaseDocument: models.BaseDocument{
			AssignmentID: uint(10),
			FileName:     "test.docx",
			FileType:     "docx",
		},
		RemoteAssignmentID: uint(10),
	}
	localDoc.ID = uint(1)

	_, err := server.GetQdrantVectors(&localDoc)
	if err != nil {
		log.Printf("err: %s", err.Error())
	}
}
