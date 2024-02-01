package getData

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"encoding/json"

	"github.com/dgraph-io/badger"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type query struct {
	RateLimit struct {
		Cost      githubv4.Int
		Limit     githubv4.Int
		NodeCount githubv4.Int
		Remaining githubv4.Int
		ResetAt   githubv4.DateTime
	}
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

func GetCommitDates(db *badger.DB, user string) (CommitDataCollection, error) {
	var cdc CommitDataCollection
	var lastFetchTime time.Time
	currentYear := time.Now().Year()
	fetchInterval := 1 * time.Hour
	dataFoundInDB := false

	// Attempt to retrieve data and last fetch time from BadgerDB
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(user))
		if err == nil {
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			err = json.Unmarshal(val, &cdc)
			if err != nil {
				return err
			}
			dataFoundInDB = true
		} else if err != badger.ErrKeyNotFound {
			return err
		}

		timeItem, err := txn.Get([]byte(user + "_lastFetchTime"))
		if err == nil {
			val, err := timeItem.ValueCopy(nil)
			if err != nil {
				return err
			}
			err = json.Unmarshal(val, &lastFetchTime)
			return err
		} else if err != badger.ErrKeyNotFound {
			return err
		}
		return nil
	})
	if err != nil {
		return CommitDataCollection{}, err
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	// If data is not found in BadgerDB, fetch all available years from GitHub
	if !dataFoundInDB {
		commitYears, err := GetCommitYears(user, client)
		if err != nil {
			return CommitDataCollection{}, err
		}
		for _, year := range commitYears {
			cdcYear, err := GetCommitDatesForYear(user, int(year), client)
			if err != nil {
				return CommitDataCollection{}, err
			}
			cdc.CommitData = append(cdc.CommitData, cdcYear.CommitData...)
		}
		lastFetchTime = time.Time{} // Reset last fetch time
	}

	// Update the current year's data if necessary
	if time.Since(lastFetchTime) >= fetchInterval || !dataFoundInDB {
		cdcYear, err := GetCommitDatesForYear(user, currentYear, client)
		if err != nil {
			return CommitDataCollection{}, err
		}

		updated := false
		for i, data := range cdc.CommitData {
			if data.Year == currentYear {
				cdc.CommitData[i] = cdcYear.CommitData[0] // Assuming cdcYear.CommitData contains current year data
				updated = true
				break
			}
		}
		if !updated {
			cdc.CommitData = append(cdc.CommitData, cdcYear.CommitData...)
		}

		// Update last fetch time
		lastFetchTime = time.Now()
	}

	// Serialize and store the updated data and last fetch time in BadgerDB
	serializedData, err := json.Marshal(cdc)
	if err != nil {
		return cdc, err
	}
	serializedTime, err := json.Marshal(lastFetchTime)
	if err != nil {
		return cdc, err
	}

	err = db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(user), serializedData)
		if err != nil {
			return err
		}
		return txn.Set([]byte(user+"_lastFetchTime"), serializedTime)
	})
	if err != nil {
		return cdc, err
	}

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

func storeCommitDataInBadger(user string, cdc CommitDataCollection) error {
	// Open the BadgerDB
	db, err := badger.Open(badger.DefaultOptions("./badgerdb.data"))
	if err != nil {
		return err
	}
	defer db.Close()

	// Serialize CommitDataCollection
	data, err := json.Marshal(cdc)
	if err != nil {
		return err
	}

	// Store the data in BadgerDB
	err = db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(user), data)
		return err
	})

	return err
}
