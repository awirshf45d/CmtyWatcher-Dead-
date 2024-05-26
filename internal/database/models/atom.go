package models

import (
	"cmtyWatcher/config"
	"cmtyWatcher/internal/database"
	"cmtyWatcher/internal/utils"
	"database/sql"
	"log"
)

// AtomFeed represents an Atom feed
type AtomFeed struct {
	Title       string      `xml:"title"`
	AuthorName  string      `xml:"author>name"`
	IdOFChannel string      `xml:"channelId"`
	Entries     []AtomEntry `xml:"entry"`
}

// AtomLink represents a link in an Atom feed
type AtomLink struct {
	Href string `xml:"href,attr"`
}

// AtomEntry represents an entry in an Atom feed
type AtomEntry struct {
	Title       string   `xml:"title"`
	ID          string   `xml:"id"`
	VideoId     string   `xml:"videoId"`
	Link        AtomLink `xml:"link"`
	IdOFChannel string   `xml:"channelId"`
	Published   string   `xml:"published"`
}

// AtomFeed methods
func (f *AtomFeed) GetItems() []AtomEntry {
	entries := make([]AtomEntry, len(f.Entries))
	entries = append(entries, f.Entries...)
	return entries
}

func (f *AtomFeed) RemoveItems(IndicesToRemove map[int]int8) {

	result := make([]AtomEntry, len(f.Entries))
	for i := range f.Entries {
		if _, found := IndicesToRemove[i]; !found {
			result = append(result, (f.Entries)[i])
		}
	}
	f.Entries = result
}

func (f *AtomFeed) ExcludeStaleFeedRecords(FeedURL string) error {
	db := database.MakeDBConnection(&config.DBconf)
	defer db.Close()
	offset := 0
	limit := 1000

	// optimize the query
	if f.Entries[0].VideoId != "" && f.IdOFChannel != "" {
		// youtube feed. it's faster than search by link.(we have lots of record, and lots of youtube channel to follow.)
		fetchYouTubeRecords := func(db *sql.DB, offset int, limit int) (map[string]int, error) {
			VideoIDsMap := make(map[string]int, 1000)

			query := "SELECT VideoID FROM FeedRecords WHERE IdOFChannel= ? LIMIT ? OFFSET ?"
			rows, err := db.Query(query, f.IdOFChannel, limit, offset)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			for rows.Next() {
				var videoId string
				if err := rows.Scan(&videoId); err != nil {
					return nil, err
				}
				VideoIDsMap[videoId] = 1
			}
			return VideoIDsMap, nil
		}

	fetchAChunkOfDataOfYoutube:
		for {
			YoutubeVideoIDsMap, err := fetchYouTubeRecords(db, offset, limit)

			if err != nil {
				log.Printf("Error While Fetching Youtube Records, %v", err.Error())
				return err
			}

			if len(YoutubeVideoIDsMap) == 0 {
				break fetchAChunkOfDataOfYoutube
			}

			// remove items which already have been added to the database.
			// a copy of items
			cEntries := make([]AtomEntry, len(f.Entries))
			copy(cEntries, f.Entries)
			EntriesToRemoveIndices := make(map[int]int8, len(f.Entries))

			for i, entry := range cEntries {
				if _, found := YoutubeVideoIDsMap[entry.VideoId]; found {
					EntriesToRemoveIndices[i] = 1
				}
			}

			f.RemoveItems(EntriesToRemoveIndices)
			offset += limit
		}
		return nil

	} else {
		// fetch items based on the link, which it cannot be null.
		// it's a little mass. but necessary
		fetchFeedRecords := func(db *sql.DB, offset int, limit int) (map[string]int, error) {
			LinksMap := make(map[string]int, 1000)

			query := "SELECT Link FROM FeedRecords WHERE SourceDomain= ? LIMIT ? OFFSET ?"

			domain, _ := utils.ExtractDomainFromURL(&FeedURL)
			rows, err := db.Query(query, domain, limit, offset)
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

			cEntries := make([]AtomEntry, len(f.Entries))
			copy(cEntries, f.Entries)
			EntriesToRemoveIndices := make(map[int]int8, len(f.Entries))

			for i, entry := range cEntries {
				if _, found := FeedsLinkMap[entry.Link.Href]; found {
					EntriesToRemoveIndices[i] = 1
				}
			}

			f.RemoveItems(EntriesToRemoveIndices)
			offset += limit
		}
		return nil

	}

}
