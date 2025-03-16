func fetchAllPapers(query string, year int, offset int, ch chan<- []Paper, wg *sync.WaitGroup) {
	defer wg.Done()
	batchSize := 100

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
			break // No more papers left
		}

		allPapers = append(allPapers, result.Data...)
		offset += batchSize
	}

	ch <- allPapers
}