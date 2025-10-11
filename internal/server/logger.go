package server

import (
	"fmt"
	"log"
	"net/http"
)

func PrintLOG(tags []string, message string) {
	var tagStr string
	for _, tag := range tags {
		tagStr += fmt.Sprintf("[%s] ", tag)
	}
	log.Printf("%s, %s", tagStr, message)
}

func PrintERROR(w http.ResponseWriter, code int, message string) {
	log.Printf("[ERROR] [%d] %s", code, message)
	http.Error(w, message, code)
}
