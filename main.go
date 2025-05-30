package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	username           = "admin"
	password           = "admin"
	url                = "https://fa68-3-108-214-79.ngrok-free.app/api/v1/dags/gini_notification_final/dagRuns"
	concurrentRequests = 100
)

func generatePayload(i int) []byte {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	dagRunID := fmt.Sprintf("manual__%s_%d", timestamp, i)

	json := fmt.Sprintf(`{
		"dag_run_id": "%s",
		"conf": {
        "eventName": "Transfer",
        "transactionId": "dab0766184b92c6c5de647bfb754ec023ebaffd7564dddd5d2a7f6357a9f7a52",
        "block_number": 912,
        "payload": {
            "from": "e3ac4b65e0bc0bfff3a88209a16cf06918c36daa",
            "to": "f8763ef1e28e3f36b86b5f7988232f88d28a6fcd",
            "value": "2000000000000000"
        }
    }
	}`, dagRunID)

	return []byte(json)
}

func postRequest(wg *sync.WaitGroup, i int) {
	defer wg.Done()

	payload := generatePayload(i)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("Request creation failed: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	req.Header.Set("Authorization", "Basic "+auth)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request %d failed: %v\n", i, err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Request %d: %s\n", i, resp.Status)
}

func main() {
	var wg sync.WaitGroup

	for i := 1; i <= concurrentRequests; i++ {
		wg.Add(1)
		go postRequest(&wg, i)
		// time.Sleep(1 * time.Second) // Optional: throttle to avoid overwhelming the server
	}

	wg.Wait()
}
