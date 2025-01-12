package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	baseURL            = "http://localhost:8080/pageLoad"
	concurrentRequests = 100 // Change this to configure the number of concurrent requests
)

type RequestPayload struct {
	TransactionName   string   `json:"transaction_name"`
	TransactionParams []string `json:"transaction_params"`
}

func makeRequest(payload RequestPayload, wg *sync.WaitGroup, id int) {
	defer wg.Done() // Notify when the goroutine completes

	// Marshal payload into JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Goroutine %d: Failed to marshal payload: %v\n", id, err)
		return
	}

	// Create HTTP POST request
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Goroutine %d: Failed to create request: %v\n", id, err)
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Goroutine %d: Request failed: %v\n", id, err)
		return
	}
	defer resp.Body.Close()

	// Log the response status
	fmt.Printf("Goroutine %d: Response status: %s\n", id, resp.Status)
}

func main() {
	// Define the payload
	// payload := RequestPayload{
	// 	TransactionName:   "AddPoints",
	// 	TransactionParams: []string{"b0f04e1e701011bc117605a1d979c5cc8e312d3a", "1"},
	// }
	payload := RequestPayload{
		TransactionName:   "GetPoints",
		TransactionParams: []string{"b0f04e1e701011bc117605a1d979c5cc8e312d3a"},
	}

	var wg sync.WaitGroup

	// Start concurrent requests
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go makeRequest(payload, &wg, i+1)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	fmt.Println("All requests completed.")
}
