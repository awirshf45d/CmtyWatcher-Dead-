package models

import (
	"cmtyWatcher/config"
	"cmtyWatcher/internal/database"
	"cmtyWatcher/internal/utils"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"
)

// RSSFeed represents an RSS feed
type RSSFeed struct {
	Channel RSSChannel `xml:"channel"`
}

// RSSChannel represents an RSS channel
type RSSChannel struct {
	Title string    `xml:"title"`
	Link  string    `xml:"link"`
	Items []RSSItem `xml:"item"`
}

// RSSItem represents an item in an RSS feed
type RSSItem struct {
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	Link        string   `xml:"link"`
	PubDate     string   `xml:"pubDate"`
	Categories  []string `xml:"category"`
	GUID        string   `xml:"guid"`
}

// RSSFeed methods
func (f *RSSFeed) GetItems() []RSSItem {
	entries := make([]RSSItem, len(f.Channel.Items))
	entries = append(entries, f.Channel.Items...)
	return entries
}

func (f *RSSFeed) RemoveItems(IndicesToRemove map[int]int8) {

	result := make([]RSSItem, len(f.Channel.Items))
	for i := range f.Channel.Items {
		if _, found := IndicesToRemove[i]; !found {
			result = append(result, (f.Channel.Items)[i])
		}
	}
	f.Channel.Items = result
}

func (f *RSSFeed) ExcludeStaleFeedRecords(FeedURL string) error {
	db := database.MakeDBConnection(&config.DBconf)
	defer db.Close()
	offset := 0
	limit := 1000

	// optimize the query

	// fetch items based on the link, which it cannot be null.
	// it's a little mass. but necessary
	fetchFeedRecords := func(db *sql.DB, offset int, limit int) (map[string]int, error) {
		LinksMap := make(map[string]int, 1000)
		var (
			rows *sql.Rows
			err  error
		)

		query := "SELECT Link FROM FeedRecords WHERE SourceDomain = ? LIMIT ? OFFSET ?"
		domain, _ := utils.ExtractDomainFromURL(&FeedURL)
		rows, err = db.Query(query, domain, limit, offset)

		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var link string
			if err := rows.Scan(&link); err != nil {
				return nil, err
			}
			LinksMap[link] = 1
		}
		return LinksMap, nil
	}

fetchAChunkOfData:
	for {
		FeedsLinkMap, err := fetchFeedRecords(db, offset, limit)

		if err != nil {
			log.Printf("Error While Fetching Feed Records, %v", err.Error())
			return err
		}

		if len(FeedsLinkMap) == 0 {
			break fetchAChunkOfData
		}

		// remove items which already have been added to the database.
		// a copy of items
		cEntries := make([]RSSItem, len(f.Channel.Items))
		copy(cEntries, f.Channel.Items)
		EntriesToRemoveIndices := make(map[int]int8, len(f.Channel.Items))

		for i, entry := range cEntries {
			if _, found := FeedsLinkMap[entry.Link]; found {
				EntriesToRemoveIndices[i] = 1
			}
		}

		f.RemoveItems(EntriesToRemoveIndices)
		offset += limit
	}
	return nil

}

func (f *RSSFeed) InsertRSSItemsToDatabase(FeedURL string) {
	db := database.MakeDBConnection(&config.DBconf)
	defer db.Close()
	layout := time.RFC1123

	var wg sync.WaitGroup
	InsertItem := func(db *sql.DB, item RSSItem) {

		defer wg.Done()
		var query string
		var err error
		// Convert PubDate to time.Time
		item.PubDate = utils.ParseTimeNiceOutput(layout, item.PubDate)
		domain, _ := utils.ExtractDomainFromURL(&FeedURL)

		if item.PubDate != "" {
			query = `
			INSERT INTO FeedRecords (Title, Link, PubDate, SourceDomain)
			VALUES (?, ?, ?, ?)`

			_, err = db.Exec(query, item.Title, item.Link, item.PubDate, domain)
		} else {
			query = `
			INSERT INTO FeedRecords (Title, Link, SourceDomain)
			VALUES (?, ?, ?)`
			_, err = db.Exec(query, item.Title, item.Link, domain)
		}

		if err != nil {
			log.Printf("error inserting RSS item: %v", err)
			return
		}
		fmt.Printf("Item Has been Added, %s\n", item.Link)

	}

	fmt.Printf("##### length of items: {%d}\n", len(f.Channel.Items))
	for _, entry := range f.Channel.Items {
		wg.Add(1)
		go InsertItem(db, entry)
	}
	wg.Wait()
}
