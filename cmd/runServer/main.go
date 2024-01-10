package main

import (
	"net/http"

	"github.com/i5heu/github-readme-commit-streak-stats/internal/generate"
	serveui "github.com/i5heu/github-readme-commit-streak-stats/internal/serveUi"
)

func main() {
	handler := serveui.ServeTemplate()
	mux := http.NewServeMux()
	mux.Handle("/", handler)
	mux.HandleFunc("/generate", generate.GenerateHandler)
	http.ListenAndServe(":8080", mux)
}
