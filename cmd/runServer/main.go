package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/i5heu/github-readme-commit-streak-stats/internal/generate"
	serveui "github.com/i5heu/github-readme-commit-streak-stats/internal/serveUi"
)

func main() {
	handler := serveui.ServeTemplate()
	mux := http.NewServeMux()
	mux.Handle("/", handler)
	mux.HandleFunc("/generate", generate.GenerateHandler)

	// checks

	if os.Getenv("GITHUB_TOKEN") == "" {
		log.Fatal("GITHUB_TOKEN environment variable not set")
	}

	fmt.Println("Starting server on port http://localhost:8080/")
	http.ListenAndServe(":8080", mux)
}
