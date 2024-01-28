package getData

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type query struct {
	User struct {
		CreatedAt               githubv4.Date
		ContributionsCollection struct {
			ContributionYears    []githubv4.Int
			ContributionCalendar struct {
				Weeks []struct {
					ContributionDays []struct {
						ContributionCount githubv4.Int
						Date              string
					}
				}
			}
		} `graphql:"contributionsCollection(from: $from, to: $to)"`
	} `graphql:"user(login: $user)"`
}

type queryYears struct {
	User struct {
		CreatedAt               githubv4.Date
		ContributionsCollection struct {
			ContributionYears []githubv4.Int
		}
	} `graphql:"user(login: $user)"`
}

type CommitDataCollection struct {
	CommitYears []int
	CommitData  []CommitData
}

type CommitData struct {
	Year  int
	Month int
	Day   int
	Count int
}

func GetCommitDates(user string) (CommitDataCollection, error) {
	err := error(nil)

	// check env for token
	if os.Getenv("GITHUB_TOKEN") == "" {
		log.Fatal("GITHUB_TOKEN environment variable not set")
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	cdc := CommitDataCollection{}
	cdc.CommitYears, err = GetCommitYears(user, client)
	if err != nil {
		log.Fatalf("Failed to get commit years: %v", err)
		return CommitDataCollection{}, err
	}

	// get commit dates for previous year
	for _, year := range cdc.CommitYears {
		cdcYear, err := GetCommitDatesForYear(user, year, client)
		if err != nil {
			log.Fatalf("Failed to get commit dates for year %d: %v", year, err)
			return CommitDataCollection{}, err
		}
		cdc.CommitData = append(cdc.CommitData, cdcYear.CommitData...)
	}

	sort.Slice(cdc.CommitData, func(i, j int) bool {
		iData := cdc.CommitData[i]
		jData := cdc.CommitData[j]

		iDate := time.Date(iData.Year, time.Month(iData.Month), iData.Day, 0, 0, 0, 0, time.UTC)
		jDate := time.Date(jData.Year, time.Month(jData.Month), jData.Day, 0, 0, 0, 0, time.UTC)

		// Return true if iDate is after jDate for descending order
		return iDate.After(jDate)
	})

	return cdc, nil
}

func GetCommitYears(githubUser string, client *githubv4.Client) ([]int, error) {
	variables := map[string]interface{}{
		"user": githubv4.String(githubUser),
	}

	var q queryYears

	err := client.Query(context.Background(), &q, variables)
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
		return []int{}, err
	}

	commitYears := make([]int, len(q.User.ContributionsCollection.ContributionYears))
	for i, year := range q.User.ContributionsCollection.ContributionYears {
		commitYears[i] = int(year)
	}

	return commitYears, nil
}

func GetCommitDatesForYear(githubUser string, year int, client *githubv4.Client) (CommitDataCollection, error) {
	fmt.Println("Getting commit dates for year", year)

	startOfYear := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	// if year is current year, endOfYear is today
	endOfYear := time.Now()
	if year != time.Now().Year() {
		endOfYear = time.Date(year, time.December, 31, 23, 59, 59, 999999999, time.UTC)
	}

	variables := map[string]interface{}{
		"user": githubv4.String(githubUser),
		"from": githubv4.DateTime{Time: startOfYear},
		"to":   githubv4.DateTime{Time: endOfYear},
	}

	var q query

	err := client.Query(context.Background(), &q, variables)
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
		return CommitDataCollection{}, err
	}

	cdc := CommitDataCollection{
		CommitYears: []int{},
		CommitData:  []CommitData{},
	}

	contributionYears := make([]string, len(q.User.ContributionsCollection.ContributionYears))
	for i, year := range q.User.ContributionsCollection.ContributionYears {
		contributionYears[i] = strconv.Itoa(int(year))
	}

	layout := "2006-01-02"

	for _, week := range q.User.ContributionsCollection.ContributionCalendar.Weeks {
		for _, day := range week.ContributionDays {

			parsedDate, err := time.Parse(layout, day.Date)
			if err != nil {
				fmt.Println("Error parsing date:", err)
				return CommitDataCollection{}, err
			}

			cdc.CommitData = append(cdc.CommitData, CommitData{
				Year:  parsedDate.Year(),
				Month: int(parsedDate.Month()),
				Day:   parsedDate.Day(),
				Count: int(day.ContributionCount),
			})
		}
	}

	return cdc, nil
}
