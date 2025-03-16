package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq"

	"github.com/Anubhav2000/research-patent-tracker/internal/models"
	"github.com/Anubhav2000/research-patent-tracker/internal/services"
)

var (
	paused       int32 = 0
	currentState models.JobState
	db           *sql.DB
)

func init() {
	// Initialize database connection
	var err error
	db, err = sql.Open("postgres", "postgres://username:password@localhost:5432/dbname?sslmode=disable")
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
}

func resumeJob() {
	if atomic.LoadInt32(&paused) == 1 {
		atomic.StoreInt32(&paused, 0)
		fmt.Println("▶️ Resuming job...")
		startJob() // This will resume from last saved progress
	} else {
		fmt.Println("✨ Job is already running")
	}
}

func startJob() {
	atomic.StoreInt32(&paused, 0)
	fmt.Println("▶️ Starting job...")

	queries := []string{}      // Add your queries here
	years := []int{2023, 2024} // Add your years here

	var wg sync.WaitGroup
	papersChan := make(chan []services.Paper)

	for _, query := range queries {
		for _, year := range years {
			wg.Add(1)
			go services.fetchAllPapers(db, query, year, papersChan, &wg)
		}
	}

	go func() {
		wg.Wait()
		close(papersChan)
	}()
}

func pauseJob() {
	atomic.StoreInt32(&paused, 1)
	fmt.Println("⏸️ Job paused. Use 'resume' to continue.")
}

func stopJob() {
	fmt.Println("🛑 Stopping job and saving progress...")
	atomic.StoreInt32(&paused, 1)
	// Progress is already saved after each batch in FetchAllPapers
	os.Exit(0)
}

func restartJob() {
	fmt.Println("🔄 Restarting job from scratch...")
	// Clear progress from database
	_, err := db.Exec("TRUNCATE fetch_progress")
	if err != nil {
		fmt.Printf("❌ Failed to clear progress: %v\n", err)
		return
	}
	startJob()
}

func saveJobState(state models.JobState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile("progress.json", data, 0644)
}

func loadJobState() (models.JobState, error) {
	data, err := os.ReadFile("progress.json")
	if err != nil {
		if os.IsNotExist(err) {
			return models.JobState{
				Queries: make(map[string]int),
			}, nil
		}
		return models.JobState{}, err
	}

	var state models.JobState
	err = json.Unmarshal(data, &state)
	return state, err
}

func runJob(state models.JobState) {
	// ... your job implementation ...
	for {
		if atomic.LoadInt32(&paused) == 1 {
			fmt.Println("⏸️ Job is paused. Waiting to resume...")
			time.Sleep(time.Second * 5)
			continue
		}
		// Your job logic here
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("❌ Usage: go run main.go <resume|pause|stop|restart>")
		return
	}

	command := os.Args[1]

	switch command {
	case "start":
		startJob()
	case "resume":
		resumeJob()
	case "pause":
		pauseJob()
	case "stop":
		stopJob()
	case "restart":
		restartJob()
	default:
		fmt.Println("❌ Invalid command. Use: resume, pause, stop, restart")
	}
}
