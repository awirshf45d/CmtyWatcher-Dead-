package utils

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"
)

func ExtractDomainFromURL(URL *string) (string, error) {
	ParsedURL, err := url.Parse(*URL)
	if err != nil {
		log.Fatal(err)
	}

	parts := strings.Split(ParsedURL.Hostname(), ".")
	if len(parts) < 2 {
		return "", errors.New("Warning: Error while Extract Domain from URL, Invalid URL.")
	}
	domain := strings.ToLower(parts[len(parts)-2] + "." + parts[len(parts)-1])
	return domain, nil

}

func RemoveItemsByIndices(ItemSlice *[]any, IndicesToRemove map[int]int8) []any {
	result := make([]any, len(*ItemSlice))
	for i := range *ItemSlice {
		if _, found := IndicesToRemove[i]; !found {
			result = append(result, (*ItemSlice)[i])
		}
	}
	return result
}

func ParseTimeNiceOutput(layout, InputTime string) string {
	// Parse the date string
	parsedTime, err := time.Parse(layout, InputTime)
	if err != nil {
		fmt.Errorf("Error parsing date:", err)
		return ""
	}

	// Format the parsed time to extract the date part
	return parsedTime.Format("2006-01-02")

}
