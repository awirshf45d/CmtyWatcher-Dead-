package feed

import (
	"cmtyWatcher/internal/database/models"
	"encoding/json"
	"encoding/xml"
	"fmt"
)

// parseRSSFeed parses an RSS feed
func ParseRSSFeed(feedData []byte) (*models.RSSFeed, error) {
	var feed models.RSSFeed
	err := xml.Unmarshal(feedData, &feed)
	if err != nil {
		return nil, err
	}
	return &feed, nil
}

// parseAtomFeed parses an Atom feed
func ParseAtomFeed(feedData []byte) (*models.AtomFeed, error) {
	var feed models.AtomFeed
	err := xml.Unmarshal(feedData, &feed)
	if err != nil {
		return nil, err
	}
	return &feed, nil
}

// parseAtomFeed parses an Atom feed
func ParsePentersterlandFeed(feedData []byte) (*models.PenLandJSONFeed, error) {
	var feed models.PenLandJSONFeed
	err := json.Unmarshal(feedData, &feed)
	if err != nil {
		fmt.Printf("\n\nERROR\n\n")
		return nil, err
	}
	return &feed, nil
}
