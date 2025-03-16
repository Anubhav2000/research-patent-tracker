package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/anubhav2000/research-patent-tracker/internal/models"
)

var (
	paused       int32 = 0
	currentState models.JobState
)

func resumeJob() {
	if atomic.LoadInt32(&paused) == 1 {
		atomic.StoreInt32(&paused, 0)
		fmt.Println("▶️ Resuming job...")
	} else {
		fmt.Println("✨ Job is already running")
	}
}

func pauseJob() {
	atomic.StoreInt32(&paused, 1)
	fmt.Println("⏸️ Job paused. Use 'resume' to continue.")
}

func stopJob() {
	fmt.Println("🛑 Stopping job and saving progress...")
	saveJobState(currentState)
	os.Exit(0)
}

func restartJob() {
	fmt.Println("🔄 Restarting job from scratch...")
	os.Remove("progress.json")
	currentState = models.JobState{
		Queries:      make(map[string]int),
		TotalFetched: 0,
	}
	runJob(currentState)
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
