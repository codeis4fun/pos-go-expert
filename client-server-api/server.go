package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "gorm.io/driver/sqlite"
)

type AwesomeAPIResponse struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

type LocalAPIResponse struct {
	Bid string `json:"bid"`
}

func getPrice(ctx context.Context, timeout int, cotacao string) (AwesomeAPIResponse, error) {
	var response AwesomeAPIResponse

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		log.Println("Could not get an answer from awesomeapi.com.br")
		return response, ctx.Err()
	default:
		log.Println("Succesfully parsed cotacao")
	}

	client := &http.Client{}

	url := fmt.Sprintf("https://economia.awesomeapi.com.br/json/last/%s", cotacao)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	if err != nil {
		log.Println("Error creating request:", err)
		return response, err
	}

	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error making request:", err)
		return response, err
	}

	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&response)

	return response, nil
}

func insertCotacao(ctx context.Context, timeout int, bid string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		log.Println("Could not insert cotacao into database!")
		return ctx.Err()
	default:
		log.Println("Data inserted successfully!")
	}

	db, err := sql.Open("sqlite3", "server.db")
	if err != nil {
		log.Println("Could not open connection to server.db:", err)
		return err
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, "INSERT INTO cotacoes (bid) VALUES (?)", bid)
	if err != nil {
		log.Println("Could not insert into cotacoes:", err)
		return err
	}

	return nil
}

func cotacaoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cotacao := "USD-BRL"
	ctx := r.Context()
	//timeout in miliseconds
	timeoutAPI := 200
	timeoutDB := 10

	price, err := getPrice(ctx, timeoutAPI, cotacao)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = insertCotacao(ctx, timeoutDB, price.USDBRL.Bid); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(price.USDBRL)

}
