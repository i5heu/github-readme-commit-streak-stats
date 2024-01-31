package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dgraph-io/badger"
	"github.com/i5heu/github-readme-commit-streak-stats/internal/generate"
	serveui "github.com/i5heu/github-readme-commit-streak-stats/internal/serveUi"
)

func main() {
	handler := serveui.ServeTemplate()
	mux := http.NewServeMux()
	mux.Handle("/", handler)

	// checks

	if os.Getenv("GITHUB_TOKEN") == "" {
		log.Fatal("GITHUB_TOKEN environment variable not set")
	}

	// Initialize BadgerDB
	db, err := badger.Open(badger.DefaultOptions("./badgerdb"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
		generate.GenerateHandler(db, w, r)
	})

	fmt.Println("Starting server on port http://localhost:8080/")
	http.ListenAndServe(":8080", mux)
}
