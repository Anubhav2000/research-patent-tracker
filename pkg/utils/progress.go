package utils

import (
	"database/sql"
	"time"
)

type FetchProgress struct {
	ID        int       `json:"id"`
	Query     string    `json:"query"`
	Year      int       `json:"year"`
	Offset    int       `json:"offset"`
	LastBatch time.Time `json:"last_batch"`
	Complete  bool      `json:"complete"`
}

func SaveProgress(db *sql.DB, progress FetchProgress) error {
	_, err := db.Exec(`
		INSERT INTO fetch_progress (query, year, offset, last_batch, complete)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (query, year) DO UPDATE
		SET offset = $3, last_batch = $4, complete = $5`,
		progress.Query, progress.Year, progress.Offset, progress.LastBatch, progress.Complete)
	return err
}

func GetProgress(db *sql.DB, query string, year int) (FetchProgress, error) {
	var progress FetchProgress
	err := db.QueryRow(`
		SELECT id, query, year, offset, last_batch, complete 
		FROM fetch_progress 
		WHERE query = $1 AND year = $2`,
		query, year).Scan(&progress.ID, &progress.Query, &progress.Year, &progress.Offset, &progress.LastBatch, &progress.Complete)

	if err == sql.ErrNoRows {
		// Return new progress if not found
		return FetchProgress{Query: query, Year: year, Offset: 0, LastBatch: time.Now(), Complete: false}, nil
	}
	return progress, err
}
