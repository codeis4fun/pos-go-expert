package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func storeCotacao(ctx context.Context, timeout int, endpoint string) error {
	var response LocalAPIResponse

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		log.Println("Could not get cotacao from endpoint")
		return ctx.Err()
	default:
		log.Println("API response content has been written to cotacao.txt")
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&response)

	// Write the response content to a file named 'cotacao.txt'
	file, err := os.Create("cotacao.txt")
	if err != nil {
		log.Printf("Error creating file: %v\n", err)
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(fmt.Sprintf("Dólar: %s", response.Bid)))
	if err != nil {
		log.Printf("Error writing to file: %v\n", err)
		return err
	}

	return nil
}
