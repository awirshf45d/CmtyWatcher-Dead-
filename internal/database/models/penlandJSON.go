package models

import (
	"cmtyWatcher/config"
	"cmtyWatcher/internal/database"
	"cmtyWatcher/internal/utils"
	"database/sql"
	"log"
)

type PenLandJSONFeed struct {
	Data []PenLandJSONFeedInnerData `json:"data"`
}
type PenLandJSONFeedInnerData struct {
	PublicationDate string             `json:"PublicationDate"`
	AddedDate       string             `json:"AddedDate"`
	Bugs            []string           `json:"Bugs"`
	Programs        []string           `json:"Programs"`
	Links           []PenlandInnerLink `json:"Links"`
}

type PenlandInnerLink struct {
	Title string `json:"Title"`
	Link  string `json:"Link"`
}

// PenLandJSONFeed methods
func (f *PenLandJSONFeed) GetItems() []PenLandJSONFeedInnerData {
	entries := make([]PenLandJSONFeedInnerData, len(f.Data))
	entries = append(entries, f.Data...)
	return entries
}

func (f *PenLandJSONFeed) RemoveItems(IndicesToRemove map[int]int8) {

	result := make([]PenLandJSONFeedInnerData, len(f.Data))
	for i := range f.Data {
		if _, found := IndicesToRemove[i]; !found {
			result = append(result, (f.Data)[i])
		}
	}
	f.Data = result
}

func (f *PenLandJSONFeed) ExcludeStaleFeedRecords(FeedURL string) error {
	db := database.MakeDBConnection(&config.DBconf)
	defer db.Close()
	offset := 0
	limit := 1000

	// fetch items based on the link, which it cannot be null.
	// it's a little mass. but necessary
	// we should set the SourceDomain of the Records to Pentesterland.land, because if we do this, in the first phase we remove lots of duplicates.
	fetchFeedRecords := func(db *sql.DB, FeedURL *string, offset int, limit int) (map[string]int, error) {
		LinksMap := make(map[string]int, 1000)

		query := "SELECT Link FROM FeedRecords WHERE SourceDomain= ? LIMIT ? OFFSET ?"
		domain, _ := utils.ExtractDomainFromURL(FeedURL)
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
		FeedsLinkMap, err := fetchFeedRecords(db, &FeedURL, offset, limit)

		if err != nil {
			log.Printf("Error While Fetching Feed Records, %v", err.Error())
			return err
		}

		if len(FeedsLinkMap) == 0 {
			break fetchAChunkOfData
		}

		// remove items which already have been added to the database.
		// a copy of items
		cEntries := make([]PenLandJSONFeedInnerData, len(f.Data))
		copy(cEntries, f.Data)
		EntriesToRemoveIndices := make(map[int]int8, len(f.Data))

		for i, entry := range cEntries {
			if _, found := FeedsLinkMap[entry.Links[0].Link]; found {
				EntriesToRemoveIndices[i] = 1
			}
		}
		f.RemoveItems(EntriesToRemoveIndices)

		offset += limit

	}
	//  we must check again, because pentesterland write-ups includes other sources like medium, so in order to prevent duplicates, we must check again. e.g. medium has lots of duplicates with pentesterland
	// phase 2, SourceDomain.
	cEntries := make([]PenLandJSONFeedInnerData, len(f.Data))
	copy(cEntries, f.Data)
	CheckedSources := make(map[string]int)
	for _, entry := range cEntries {
		offset = 0
		limit = 1000
		domain, err := utils.ExtractDomainFromURL(&entry.Links[0].Link)
		if err != nil {
			continue
		} else if _, found := CheckedSources[domain]; found {
			continue
		}
		CheckedSources[domain] = 1
	fetchAChunkOfDataPhase2:
		for {

			FeedsLinkMap, err := fetchFeedRecords(db, &domain, offset, limit)

			if err != nil {
				log.Printf("Error While Fetching Feed Records, %v", err.Error())
				return err
			}

			if len(FeedsLinkMap) == 0 {
				break fetchAChunkOfDataPhase2
			}

			// remove items which already have been added to the database.
			// a copy of items

			for i, entry := range cEntries {
				if _, found := FeedsLinkMap[entry.Links[0].Link]; found {
					f.Data = append(f.Data[:i], f.Data[i+1:]...)
				}
			}
			offset += limit

		}
	}

	return nil

}
