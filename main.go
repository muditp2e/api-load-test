package main

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

const (
	URL     = "https://kwala-web2-loadnet.p2eppl.com/user/getBalance/kwl-6cd5851c9c6f98bfe40febdeb63e95e9a0677a19-cc"
	N       = 50
	Timeout = 15 * time.Second
)

type Result struct {
	Latency time.Duration
	OK      bool
	Status  int
}

func oneRequest(client *http.Client) Result {
	start := time.Now()
	resp, err := client.Get(URL)
	lat := time.Since(start)

	if err != nil {
		return Result{Latency: lat, OK: false, Status: 0}
	}
	defer resp.Body.Close()

	ok := resp.StatusCode >= 200 && resp.StatusCode < 300
	return Result{Latency: lat, OK: ok, Status: resp.StatusCode}
}

func summarize(name string, results []Result) {
	if len(results) == 0 {
		fmt.Println(name, "no results")
		return
	}

	latencies := make([]time.Duration, 0, len(results))
	okCount := 0
	statusCounts := map[int]int{}

	for _, r := range results {
		latencies = append(latencies, r.Latency)
		if r.OK {
			okCount++
		} else {
			statusCounts[r.Status]++
		}
	}

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	min := latencies[0]
	max := latencies[len(latencies)-1]
	var sum time.Duration
	for _, l := range latencies {
		sum += l
	}
	avg := time.Duration(int64(sum) / int64(len(latencies)))

	fmt.Printf("\n=== %s ===\n", name)
	fmt.Printf("Total: %d | Success: %d | Errors: %d\n", len(results), okCount, len(results)-okCount)
	fmt.Printf("Latency: min=%s avg=%s max=%s\n", min, avg, max)

	if len(statusCounts) > 0 {
		fmt.Printf("Non-2xx status counts: %+v\n", statusCounts)
	}
}

func sequential(client *http.Client) {
	results := make([]Result, 0, N)
	for i := 0; i < N; i++ {
		results = append(results, oneRequest(client))
	}
	summarize("A) Sequential (50x)", results)
}

func concurrent(client *http.Client) {
	results := make([]Result, N)
	var wg sync.WaitGroup
	wg.Add(N)

	for i := 0; i < N; i++ {
		i := i
		go func() {
			defer wg.Done()
			results[i] = oneRequest(client)
		}()
	}
	wg.Wait()
	summarize("B) Concurrent (50x)", results)
}

func main() {
	client := &http.Client{Timeout: Timeout}

	fmt.Println("Target:", URL)
	sequential(client)
	concurrent(client)
}
