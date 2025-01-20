package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	baseURL            = "https://alpha-wallet-api.kalp.studio/wallet/sign-kalp-transaction-for-test"
	concurrentRequests = 10 // Change this to configure the number of concurrent requests
)

type RequestPayload struct {
	EnrollmentID      string   `json:"enrollmentID"`
	ChannelName       string   `json:"channelName"`
	ChainCodeName     string   `json:"chainCodeName"`
	TransactionName   string   `json:"transactionName"`
	TransactionParams []string `json:"transactionParams"`
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
	// id, err := generateHexString(22)
	// if err != nil {
	// 	log.Fatalf("Failed to generate random ID: %v", err)
	// }
	// Create a JSON message
	payload := RequestPayload{
		EnrollmentID:      "b0f04e1e701011bc117605a1d979c5cc8e312d3a",
		ChannelName:       "kalp",
		ChainCodeName:     "rpccu",
		TransactionName:   "AddPoints",
		TransactionParams: []string{"b0f04e1e701011bc117605a1d979c5cc8e312d3a", "1"},
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

func generateHexString(length int) (string, error) {
	if length%2 != 0 {
		return "", fmt.Errorf("length must be even to represent bytes as hexadecimal")
	}

	// Calculate the number of bytes needed
	byteLength := length / 2
	bytes := make([]byte, byteLength)

	// Fill the byte slice with random data
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode the bytes to hexadecimal
	return hex.EncodeToString(bytes), nil
}
