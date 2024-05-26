package models

type FeedData interface {
	GetItems() []string
	RemoveItems(index int)
	ExcludeStaleFeedRecords()
}
