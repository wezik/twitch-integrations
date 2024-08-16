package database

import (
	"database/sql"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbPath        = "./squeel.db"
	versionQuery  = "SELECT COUNT(*) FROM queries"
	versionInsert = "INSERT INTO queries (id, query) VALUES (?, ?)"
)

func Init() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	log.Println("Database initialized")
	err = runQueries(db)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func runQueries(db *sql.DB) error {
	var count int = 0
	err := db.QueryRow(versionQuery).Scan(&count)
	if err != nil {
		count = 0
	}

	if count == len(queries) {
		log.Println("Database is up to date")
		return nil
	}

	log.Println("Updating database...")
	for i := count; i < len(queries); i++ {
		query := strings.ReplaceAll(queries[i], "\n", "")
		query = strings.Join(strings.Fields(query), " ")
		log.Println("Running query", query)
		_, err := db.Exec(query)
		if err != nil {
			return err
		}
		_, err = db.Exec(versionInsert, i, query)
		if err != nil {
			return err
		}
	}

	log.Println("Database is up to date")
	return nil
}

var queries = []string{
	`CREATE TABLE IF NOT EXISTS queries (
		id INTEGER PRIMARY KEY,
		query TEXT NOT NULL
	);`,
	`CREATE TABLE IF NOT EXISTS tokens (
		id TEXT PRIMARY KEY,
		token TEXT NOT NULL
	);`,
}
