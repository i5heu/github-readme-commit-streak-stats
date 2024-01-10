package getdata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Event struct {
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

func GetCommitDates(user string) ([]time.Time, error) {
	var commitDates []time.Time
	page := 1
	totalEvents := 0

	for {
		resp, err := http.Get(fmt.Sprintf("https://api.github.com/users/%s/events?page=%d", user, page))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var events []Event
		if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
			return nil, err
		}

		// If no events are returned, we've reached the end of the pages
		if len(events) == 0 {
			break
		}

		for _, event := range events {
			if event.Type == "PushEvent" {
				commitDates = append(commitDates, event.CreatedAt)
			}
			totalEvents++
		}

		page++
	}

	fmt.Printf("Total events: %d\n", totalEvents)

	return commitDates, nil
}
