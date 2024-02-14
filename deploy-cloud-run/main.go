package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const cepLength = 8

type (
	temperature float64
	tempC       temperature
	tempF       temperature
	tempK       temperature
)

type weatherAPI struct {
	apiKey  string
	client  *http.Client
	Current struct {
		TempC tempC `json:"temp_c"`
		TempF tempF `json:"temp_f"`
		TempK tempK `json:"temp_k"`
	} `json:"current"`
}

func (w *weatherAPI) tempC() tempC {
	return w.Current.TempC
}

func (w *weatherAPI) tempCToF() tempF {
	return tempF(w.Current.TempC*9/5 + 32)
}

func (w *weatherAPI) tempCToK() tempK {
	return tempK(w.Current.TempC + 273)
}

func (w *weatherAPI) getWeather(localidade string) (*weatherAPI, error) {
	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", w.apiKey, url.QueryEscape(localidade))

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, err
	}

	return w, nil
}

type response struct {
	TempC tempC `json:"temp_C"`
	TempF tempF `json:"temp_F"`
	TempK tempK `json:"temp_K"`
}

func (r *response) setTempC(temp tempC) {
	r.TempC = temp
}

func (r *response) setTempF(temp tempF) {
	r.TempF = temp
}

func (r *response) setTempK(temp tempK) {
	r.TempK = temp
}

type viaCEP struct {
	client     *http.Client
	Localidade string `json:"localidade"`
}

func (v *viaCEP) getLocalidade(cep string) string {
	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return ""
	}

	req.Header.Set("Accept", "application/json")
	resp, err := v.client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var viaCEP viaCEP
	if err := json.NewDecoder(resp.Body).Decode(&viaCEP); err != nil {
		return ""
	}

	return viaCEP.Localidade
}

func main() {
	apiKey := os.Getenv("WEATHER_API_KEY")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := new(http.Client)
	client.Transport = tr

	defer client.CloseIdleConnections()

	viaCEP := &viaCEP{client: client}
	weatherAPI := &weatherAPI{apiKey: apiKey, client: client}

	http.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
		cep := r.URL.Query().Get("cep")
		if strings.Count(cep, "")-1 != cepLength {
			http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
			return
		}

		localidade := viaCEP.getLocalidade(cep)
		log.Println("localidade: ", localidade)
		if localidade == "" {
			http.Error(w, "can not find zipcode", http.StatusNotFound)
			return
		}

		weather, err := weatherAPI.getWeather(localidade)
		log.Println("weather: ", weather)
		if err != nil {
			fmt.Println(err)
			return
		}

		response := new(response)
		response.setTempC(weather.tempC())
		response.setTempF(weather.tempCToF())
		response.setTempK(weather.tempCToK())

		log.Println("response from /weather")
		log.Println(response)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	http.ListenAndServe(":8080", nil)
}
