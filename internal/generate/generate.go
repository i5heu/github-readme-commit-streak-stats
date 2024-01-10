package generate

import (
	"fmt"
	"net/http"

	getdata "github.com/i5heu/github-readme-commit-streak-stats/internal/getData"
)

func GenerateHandler(w http.ResponseWriter, r *http.Request) {
	githubUser := r.URL.Query().Get("githubUser")
	mode := r.URL.Query().Get("mode")
	strictness := r.URL.Query().Get("strictness")

	fmt.Println(githubUser, mode, strictness)

	commitDates, err := getdata.GetCommitDates(githubUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, date := range commitDates {
		w.Write([]byte(date.String() + "\n"))
	}
}
