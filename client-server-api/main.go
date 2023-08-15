package main

import (
	"context"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/cotacao", cotacaoHandler)
	go http.ListenAndServe(":8080", nil)
	timeoutStore := 300
	ctx := context.Background()
	if err := storeCotacao(ctx, timeoutStore, "http://localhost:8080/cotacao"); err != nil {
		log.Println("Could not send request to local endpoint")
	}
}
