package database

import (
	"cmtyWatcher/config"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func MakeDBConnection(DBconf *config.DBConfig) *sql.DB {
	// fetch feed urls from the database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", DBconf.DBUsername, DBconf.DBPassword, DBconf.DBHost, DBconf.DBPort, DBconf.DBName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error connecting to database: %v\n", err)
	}
	return db
}

func FetchFeedURLsFromDB() ([]string, error) {
	db := MakeDBConnection(&config.DBconf)
	defer db.Close()

	var urls []string
	query := "SELECT url FROM feed_urls;"
	rows, err := db.Query(query)

	if err != nil {
		return nil, fmt.Errorf("error fetching feed URLs: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		urls = append(urls, url)
	}

	return urls, nil
}
