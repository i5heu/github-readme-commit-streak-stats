package generate

import (
	"fmt"
	"image/jpeg"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/fogleman/gg"
	"github.com/i5heu/github-readme-commit-streak-stats/internal/getData"
)

func GenerateHandler(db *badger.DB, w http.ResponseWriter, r *http.Request) {
	githubUser, bonusDayEvery, err := getUserInputAndSanitizeIt(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cdc, err := getData.GetCommitDates(db, githubUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalCommits := 0
	for _, cd := range cdc.CommitData {
		totalCommits += cd.Count
	}

	// Calculate streaks
	currentStreak, longestStreak, bonusDays := CalculateStreaks(cdc.CommitData, bonusDayEvery)

	// Start drawing
	const W = 700
	const H = 195
	dc := gg.NewContext(W, H)

	// Set the background color, e.g., black
	dc.SetRGB(0, 0, 0)
	dc.Clear()

	// Set the text color, e.g., white
	dc.SetRGB(1, 1, 1)

	// Load a font face at double the size of the default (assuming default is 12)
	if err := dc.LoadFontFace("./fonts/Roboto/Roboto-Regular.ttf", 35); err != nil {
		log.Println("Failed to load font:", err)
		http.Error(w, "Failed to load font", http.StatusInternalServerError)
		return
	}

	heightOfLabel := 4.0
	heightOfNumber := 2.1

	dc.DrawStringAnchored(fmt.Sprintf("%d", totalCommits), W*0.18, H/2.1, 0.5, 0.5)
	dc.DrawStringAnchored(fmt.Sprintf("%d  Days", currentStreak), W*0.5, H/heightOfNumber, 0.5, 0.5)
	dc.DrawStringAnchored(fmt.Sprintf("%d  Days", longestStreak), W*0.82, H/heightOfNumber, 0.5, 0.5)

	if err := dc.LoadFontFace("./fonts/Roboto/Roboto-Regular.ttf", 23); err != nil {
		log.Println("Failed to load font:", err)
		http.Error(w, "Failed to load font", http.StatusInternalServerError)
		return
	}

	dc.DrawStringAnchored(fmt.Sprintf("Total Contributions"), W*0.18, H/4, 0.5, 0.5)
	dc.DrawStringAnchored(fmt.Sprintf("Current Streak"), W*0.5, H/heightOfLabel, 0.5, 0.5)
	dc.DrawStringAnchored(fmt.Sprintf("Longest Streak"), W*0.82, H/heightOfLabel, 0.5, 0.5)

	if err := dc.LoadFontFace("./fonts/Roboto/Roboto-Regular.ttf", 16); err != nil {
		log.Println("Failed to load font:", err)
		http.Error(w, "Failed to load font", http.StatusInternalServerError)
		return
	}

	heightOfLabel = 1.35
	heightOfNumber = 1.15

	dc.DrawStringAnchored(fmt.Sprintf("Grace Days Left"), W*0.5, H/heightOfLabel, 0.5, 0.5)
	dc.DrawStringAnchored(fmt.Sprintf("%d Days", bonusDays), W*0.5, H/heightOfNumber, 0.5, 0.5)

	dc.DrawStringAnchored(fmt.Sprintf("Grace Day Every"), W*0.82, H/heightOfLabel, 0.5, 0.5)
	dc.DrawStringAnchored(fmt.Sprintf("%d Consecutive Commit Days", bonusDayEvery), W*0.82, H/heightOfNumber, 0.5, 0.5)

	// Finish drawing
	dc.Stroke()

	// Encode as PNG
	// err = dc.EncodePNG(w)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	img := dc.Image()

	// Encode as JPEG
	quality := 80
	err = jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

// Assuming sortCommitData function sorts commitData in ascending order by date
func sortCommitData(commitData []getData.CommitData) []getData.CommitData {
	sort.Slice(commitData, func(i, j int) bool {
		iDate := time.Date(commitData[i].Year, time.Month(commitData[i].Month), commitData[i].Day, 0, 0, 0, 0, time.UTC)
		jDate := time.Date(commitData[j].Year, time.Month(commitData[j].Month), commitData[j].Day, 0, 0, 0, 0, time.UTC)
		return iDate.Before(jDate)
	})
	return commitData
}

// CalculateCurrentStreak calculates the current streak
// returns current streak, longest streak, bonus days left
func CalculateStreaks(commitData []getData.CommitData, bonusDayEvery int) (int, int, int) {
	commitData = sortCommitData(commitData)

	type StreakData struct {
		Streak                   int
		BonusDays                int
		EligibleContributionDays int
	}

	streakIndex := 0
	streakDataSlice := []StreakData{{Streak: 0, BonusDays: 0, EligibleContributionDays: 0}}

	for _, cd := range commitData {
		// Need to make sure we have an element to access
		if streakIndex >= len(streakDataSlice) {
			streakDataSlice = append(streakDataSlice, StreakData{})
		}

		if cd.Count > 0 || streakDataSlice[streakIndex].BonusDays > 0 || cd == commitData[len(commitData)-1] {

			if streakDataSlice[streakIndex].EligibleContributionDays == bonusDayEvery {
				streakDataSlice[streakIndex].BonusDays++
				streakDataSlice[streakIndex].EligibleContributionDays = 0
			}

			if cd.Count == 0 {
				streakDataSlice[streakIndex].EligibleContributionDays = 0
			}

			if streakDataSlice[streakIndex].BonusDays > 0 && cd.Count == 0 {
				streakDataSlice[streakIndex].Streak++
				streakDataSlice[streakIndex].BonusDays--
				streakDataSlice[streakIndex].EligibleContributionDays = 0
			}

			if cd.Count > 0 {
				streakDataSlice[streakIndex].EligibleContributionDays++
				streakDataSlice[streakIndex].Streak++
			}

		} else {
			streakIndex++
			streakDataSlice = append(streakDataSlice, StreakData{})
		}
	}

	// Handle case where there's no commit data
	if len(streakDataSlice) == 0 {
		return 0, 0, 0
	}

	// Calculate longest streak
	longestStreak := 0
	for _, sd := range streakDataSlice {
		if sd.Streak > longestStreak {
			longestStreak = sd.Streak
		}
	}

	return streakDataSlice[len(streakDataSlice)-1].Streak, longestStreak, streakDataSlice[len(streakDataSlice)-1].BonusDays
}

func getUserInputAndSanitizeIt(r *http.Request) (string, int, error) {
	githubUser := r.URL.Query().Get("githubUser")
	bonusDayEvery := r.URL.Query().Get("bonusDayEvery")
	if bonusDayEvery == "" {
		bonusDayEvery = "3"
	}

	// remove everything except alphanumeric character, underscore, and dash
	githubUser = regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(githubUser, "")

	// remove everything except numbers
	bonusDayEvery = regexp.MustCompile(`[^0-9]`).ReplaceAllString(bonusDayEvery, "")
	bonusDayEveryInt, err := strconv.Atoi(bonusDayEvery)

	return githubUser, bonusDayEveryInt, err
}
