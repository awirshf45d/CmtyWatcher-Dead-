package main

import (
	"cmtyWatcher/config"
	"cmtyWatcher/internal/database"
	"cmtyWatcher/internal/feed"
	httpclient "cmtyWatcher/internal/http-client"
	"fmt"
	"log"
	"strings"
	"sync" // Import sync package for WaitGroup
)

func main() {

	urls, err := database.FetchFeedURLsFromDB()
	if err != nil {
		log.Fatalf("Error while fetching feed URLs from Database: %v\n", err.Error())
	}

	client := httpclient.NewClient(config.Proxies[0], "")
	client_WithoutProxy := httpclient.NewClient("", "")

	// Create a WaitGroup
	var wg sync.WaitGroup
	for _, feedURL := range urls {
		// Increment the WaitGroup counter
		wg.Add(1)

		go func(feedURL string) {
			defer wg.Done()
			responseBody, err := client.FetchURL(feedURL)
			if err != nil {
				log.Printf("Error fetching URL %v\nProxyURL: %v\n\n", err.Error(), config.Proxies[0])
				responseBody, err = client_WithoutProxy.FetchURL(feedURL)
				if err != nil {
					log.Printf("Error fetching URL\n %v\nProxyURL: %v\n", err.Error(), "Nothing.")
					return
				}
			}

			feedType := feed.DetectFeedType(strings.NewReader(responseBody))
			switch feedType {
			case feed.FeedTypeAtom:
				atomFeed, err := feed.ParseAtomFeed([]byte(responseBody))
				if err != nil {
					log.Fatalf("Error while parsing AtomFees, %v", err.Error())
				}
				if len(atomFeed.Entries) == 0 {
					return
				}
				fmt.Println("ATOM")
				err = atomFeed.ExcludeStaleFeedRecords(feedURL)
				if err != nil {
					log.Printf("Error Excluding Stale Data: %v", err.Error())
					return
				}
			case feed.FeedTypeRSS:
				rssfeed, err := feed.ParseRSSFeed([]byte(responseBody))
				if err != nil {
					log.Fatalf("Error while parsing RSS, %v", err.Error())
				}
				if len(rssfeed.Channel.Items) == 0 {
					return
				}
				err = rssfeed.ExcludeStaleFeedRecords(feedURL)
				if err != nil {
					log.Printf("Error Excluding Stale Data: %v", err.Error())
					return
				}
				rssfeed.InsertRSSItemsToDatabase(feedURL)
				fmt.Println("RSS")

			case feed.FeedTypePenLandJSON:
				penland, err := feed.ParsePentersterlandFeed([]byte(responseBody))
				if err != nil {
					log.Fatalf("Error while parsing PENLAND, %v", err.Error())
				}
				if len(penland.Data) == 0 {
					return
				}
				err = penland.ExcludeStaleFeedRecords(feedURL)
				if err != nil {
					log.Printf("Error Excluding Stale Data: %v", err.Error())
					return
				}

			}
			if err != nil {
				log.Fatalf("Failed to parse the response body: %v", err)
			}
		}(feedURL)

	}
	// Wait for all goroutines to finish
	wg.Wait()
}
