rss-matrix-notifier/
├── cmd/
│   └── main.go          # Entry point of the application
├── config/
│   └── config.go        # Configuration handling (e.g., database, RSS feed URLs, Matrix credentials)
├── internal/
│   ├── rss/
│   │   ├── rss.go       # Functions to fetch and parse RSS feeds
│   │   └── rss_test.go  # Unit tests for RSS fetching
│   ├── database/
│   │   ├── db.go        # Database connection and operations (CRUD)
│   │   └── db_test.go   # Unit tests for database operations
│   ├── comparer/
│   │   ├── compare.go   # Logic to compare RSS feed data with local database records
│   │   └── compare_test.go # Unit tests for comparison logic
│   ├── notifier/
│   │   ├── matrix.go    # Functions to push notifications to Matrix
│   │   └── matrix_test.go # Unit tests for Matrix notifications
│   └── models/
│       └── models.go    # Data models (structs) for RSS items and database records
├── scripts/
│   └── migrate.sh       # Database migration scripts
├── vendor/              # Dependencies (if vendoring is used)
├── go.mod               # Go module file
├── go.sum               # Go dependencies checksum
├── README.md            # Project documentation
└── Makefile             # Makefile for build automation
# CmtryWatcher
