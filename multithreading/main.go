package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type ApiCep struct {
	Cep        string `json:"code,omitempty"`
	Uf         string `json:"state,omitempty"`
	Localidade string `json:"city,omitempty"`
	Bairro     string `json:"district,omitempty"`
	Endereco   string `json:"address,omitempty"`
}

type ViaCep struct {
	Cep         string `json:"cep,omitempty"`
	Uf          string `json:"uf,omitempty"`
	Localidade  string `json:"localidade,omitempty"`
	Bairro      string `json:"bairro,omitempty"`
	Logradouro  string `json:"logradouro,omitempty"`
	Complemento string `json:"complemento,omitempty"`
}

type BrasilApi struct {
	Cep        string `json:"cep,omitempty"`
	Uf         string `json:"state,omitempty"`
	Localidade string `json:"city,omitempty"`
	Bairro     string `json:"neighborhood,omitempty"`
	Endereco   string `json:"street,omitempty"`
}

type Response struct {
	Endpoint string
	Data     interface{} `json:"-"`
	Status   int
}

func setEndpoint(endpoint, cep string) string {
	endpoints := map[string]string{
		"ApiCep":    fmt.Sprintf("https://cdn.apicep.com/file/apicep/%s.json", cep),
		"ViaCep":    fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep),
		"BrasilApi": fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep),
	}

	return endpoints[endpoint]
}

func sendRequest(ctx context.Context, endpoint string, cep string, ch chan<- Response, wg *sync.WaitGroup) {
	defer wg.Done()

	var response Response
	url := setEndpoint(endpoint, cep)

	client := &http.Client{Timeout: 1 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request for endpoint %s: %v\n", endpoint, err)
		ch <- response
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request for endpoint %s: %v\n", endpoint, err)
		ch <- response
		return
	}
	defer resp.Body.Close()

	// Determine the appropriate struct type based on the endpoint
	var data interface{}
	switch endpoint {
	case "ApiCep":
		data = &ApiCep{}
	case "ViaCep":
		data = &ViaCep{}
	case "BrasilApi":
		data = &BrasilApi{}
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		fmt.Printf("Error decoding response for endpoint %s: %v\n", endpoint, err)
		ch <- response
		return
	}

	response.Endpoint = endpoint
	response.Data = data
	response.Status = resp.StatusCode
	ch <- response
}

func main() {
	var wg sync.WaitGroup

	cep := "22241-330"

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan Response, 3)

	wg.Add(3)
	go sendRequest(ctx, "ApiCep", cep, ch, &wg)
	go sendRequest(ctx, "ViaCep", cep, ch, &wg)
	go sendRequest(ctx, "BrasilApi", cep, ch, &wg)

	// Wait for the first response and cancel the other request
	go func() {
		response := <-ch
		fmt.Printf("The endpoint %s returned status code %d and the response is: %+v\n", response.Endpoint, response.Status, response.Data)
		cancel()
	}()

	wg.Wait() // Wait for both requests to finish
	close(ch)
}
