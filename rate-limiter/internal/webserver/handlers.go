package webserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/codeis4fun/pos-go-expert/rate-limiter/pkg/repository"
	"github.com/go-chi/render"
)

func (ws *WebServer) RateLimiterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.RemoteAddr
		if r.Header.Get("API_KEY") != "" {
			key = r.Header.Get("API_KEY")
		}

		config, err := ws.Repository.GetConfigByConfigValue(key)
		if err != nil && err.Error() != "record not found" {
			log.Println("failed to get config by valeu:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response := []byte("You are free to acces the endpoint\n")
		if r.Header.Get("API_KEY") != "" {
			response = append(response, fmt.Sprintf("Your API key is: %s\n", r.Header.Get("API_KEY"))...)
		}
		response = append(response, fmt.Sprintf("Your IP address is: %s\n", r.RemoteAddr)...)
		if config.MaxRequest == 0 {
			config.MaxRequest = ws.Config.RateLimitMaxRequests
		}
		if config.BlockTime == 0 {
			config.BlockTime = ws.Config.RateLimitBlockTime
		}
		response = append(response, fmt.Sprintf("Your request limit is: %d\n", config.MaxRequest)...)
		response = append(response, fmt.Sprintf("Your block time is: %d\n", config.BlockTime)...)
		if config.LimitType == "" {
			config.LimitType = "GLOBAL"
		}
		response = append(response, fmt.Sprintf("Your limit type is: %s\n", config.LimitType)...)
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	}
}

func (ws *WebServer) GetConfigByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		config, err := ws.Repository.GetConfigByID(id)
		if err != nil {
			if err.Error() == "record not found" {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, config)
	}
}

func (ws *WebServer) GetAllConfigs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		configs, err := ws.Repository.GetAllConfigs()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, configs)
	}
}

func (ws *WebServer) CreateConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var config repository.RateLimitConfig

		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := ws.Repository.CreateConfig(config); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func (ws *WebServer) UpdateConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var config repository.RateLimitConfig

		id := r.URL.Query().Get("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		idInt, err := strconv.Atoi(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		config.Id = &idInt

		if err := ws.Repository.UpdateConfig(config); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (ws *WebServer) DeleteConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := ws.Repository.DeleteConfig(id); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
