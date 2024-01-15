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

	query, err := getData.GetCommitDates(githubUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, week := range query.User.ContributionsCollection.ContributionCalendar.Weeks {
		for _, day := range week.ContributionDays {
			fmt.Printf("On %s, user %s made %d contributions\n", day.Date, githubUser, day.ContributionCount)
		}
	}
}
