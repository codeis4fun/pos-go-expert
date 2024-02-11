package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type stressReport map[int]*atomic.Uint64

func newStressReport() *stressReport {
	sr := make(stressReport)
	// Add status code 200 to the report in case the stress report only returns other status codes
	sr[http.StatusOK] = new(atomic.Uint64)
	return &sr
}

func (sr *stressReport) increment(code int) {
	if _, ok := (*sr)[code]; !ok {
		(*sr)[code] = new(atomic.Uint64)
	}

	(*sr)[code].Add(1)
}

func (sr *stressReport) String() string {
	var result string
	for code, count := range *sr {
		result += fmt.Sprintf("Status code %d: %d\n", code, count.Load())
	}
	return result
}

func makeRequest(ctx context.Context, wg *sync.WaitGroup, sr *stressReport, url string) {
	defer wg.Done()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Println(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	sr.increment(resp.StatusCode)
}

func main() {
	var (
		ctx                   context.Context
		wg                    *sync.WaitGroup
		url                   string
		requests, concurrency int
		totalRequests         *atomic.Uint64
	)

	flag.StringVar(&url, "url", "https://www.google.com.br", "URL to send requests to")
	flag.IntVar(&requests, "requests", 100, "Number of requests to send")
	flag.IntVar(&concurrency, "concurrency", 10, "Number of requests to send concurrently")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	report := newStressReport()
	wg = new(sync.WaitGroup)
	totalRequests = new(atomic.Uint64)

	now := time.Now()
	for int(totalRequests.Load()) < requests {
		remainingRequests := requests - int(totalRequests.Load())
		numRequests := int(math.Min(float64(concurrency), float64(remainingRequests)))
		wg.Add(numRequests)
		for j := 1; j <= numRequests; j++ {
			go func() {
				makeRequest(ctx, wg, report, url)
				totalRequests.Add(1) // Increment the total requests counter
			}()
		}
		wg.Wait()
	}
	fmt.Println("Report:")
	fmt.Println(fmt.Sprintf("All requests finished in %s", time.Since(now)))
	fmt.Println(fmt.Sprintf("Total requests: %d", totalRequests.Load()))
	fmt.Println(report)
}
