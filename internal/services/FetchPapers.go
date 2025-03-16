package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/Anubhav2000/research-patent-tracker/pkg/utils"
)

type Paper struct {
	Title     string   `json:"title"`
	Authors   []string `json:"authors"`
	Year      int      `json:"year"`
	Citations int      `json:"citations"`
	URL       string   `json:"url"`
}

func fetchAllPapers(db *sql.DB, query string, year int, ch chan<- []Paper, wg *sync.WaitGroup) {
	defer wg.Done()
	batchSize := 100

	// Get existing progress
	progress, err := utils.GetProgress(db, query, year)
	if err != nil {
		fmt.Printf("❌ Failed to get progress: %v\n", err)
		return
	}

	// Start from last saved offset
	offset := progress.Offset
	var allPapers []Paper

	for {
		url := fmt.Sprintf("%s?query=%s&year=%d&offset=%d&limit=%d", apiURL, query, year, offset, batchSize)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("❌ API Request Failed:", err)
			break
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		var result struct {
			Data []Paper `json:"data"`
		}
		json.Unmarshal(body, &result)

		if len(result.Data) == 0 {
			// Mark as complete when no more papers
			progress.Complete = true
			utils.SaveProgress(db, progress)
			break
		}

		// Store papers in PostgreSQL
		if err := storePapers(db, result.Data); err != nil {
			fmt.Printf("❌ Failed to store papers: %v\n", err)
			break
		}

		allPapers = append(allPapers, result.Data...)
		offset += batchSize

		// Update progress after each batch
		progress.Offset = offset
		progress.LastBatch = time.Now()
		if err := utils.SaveProgress(db, progress); err != nil {
			fmt.Printf("❌ Failed to save progress: %v\n", err)
		}
	}

	ch <- allPapers
}

func storePapers(db *sql.DB, papers []Paper) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO papers (title, authors, year, citations, url)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (url) DO UPDATE
		SET title = $1, authors = $2, year = $3, citations = $4`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, paper := range papers {
		authors, _ := json.Marshal(paper.Authors)
		_, err = stmt.Exec(paper.Title, authors, paper.Year, paper.Citations, paper.URL)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
