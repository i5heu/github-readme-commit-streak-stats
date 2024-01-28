package generate

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fogleman/gg"
	"github.com/i5heu/github-readme-commit-streak-stats/internal/getData"
)

func GenerateHandler(w http.ResponseWriter, r *http.Request) {
	githubUser := r.URL.Query().Get("githubUser")

	cdc, err := getData.GetCommitDates(githubUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalCommits := 0
	for _, cd := range cdc.CommitData {
		totalCommits += cd.Count
	}

	firstCommit := cdc.CommitData[len(cdc.CommitData)-1]
	firstCommitDate := fmt.Sprintf("%d-%02d-%02d", firstCommit.Year, firstCommit.Month, firstCommit.Day)

	const W = 400
	const H = 200
	dc := gg.NewContext(W, H)

	dc.SetRGB(1, 1, 1)
	dc.Clear()

	dc.SetRGB(0, 0, 0)
	dc.DrawStringAnchored("Total commits from "+firstCommitDate+" - Present: "+strconv.Itoa(totalCommits), W/2, H/2, 0.5, 0.5)
	dc.Stroke()

	err = dc.EncodePNG(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
