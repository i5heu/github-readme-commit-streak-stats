package generate

import (
	"fmt"
	"net/http"

	"github.com/i5heu/github-readme-commit-streak-stats/internal/getData"
)

func GenerateHandler(w http.ResponseWriter, r *http.Request) {
	githubUser := r.URL.Query().Get("githubUser")
	mode := r.URL.Query().Get("mode")
	strictness := r.URL.Query().Get("strictness")

	fmt.Println(githubUser, mode, strictness)

	cdc, err := getData.GetCommitDates(githubUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, cd := range cdc.CommitData {
		w.Write([]byte(fmt.Sprintf("%d-%d-%d %d\n", cd.Year, cd.Month, cd.Day, cd.Count)))
	}
}
