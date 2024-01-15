package getData

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"

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
		}
	} `graphql:"user(login: $user)"`
}

func GetCommitDates(user string) (query, error) {

	// check env for token
	if os.Getenv("GITHUB_TOKEN") == "" {
		log.Fatal("GITHUB_TOKEN environment variable not set")
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := githubv4.NewClient(httpClient)

	variables := map[string]interface{}{
		"user": githubv4.String(user),
	}

	var q query

	err := client.Query(context.Background(), &q, variables)
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
		return q, err
	}

	contributionYears := make([]string, len(q.User.ContributionsCollection.ContributionYears))
	for i, year := range q.User.ContributionsCollection.ContributionYears {
		contributionYears[i] = strconv.Itoa(int(year))
	}

	log.Printf("User %s has contributed in the following years: %s", user, strings.Join(contributionYears, ", "))

	return q, nil
}
