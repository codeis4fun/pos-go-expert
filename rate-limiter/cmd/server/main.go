package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/codeis4fun/pos-go-expert/rate-limiter/config"
	"github.com/codeis4fun/pos-go-expert/rate-limiter/internal/httprate"
	"github.com/codeis4fun/pos-go-expert/rate-limiter/internal/webserver"
	"github.com/codeis4fun/pos-go-expert/rate-limiter/pkg/cache"
	"github.com/codeis4fun/pos-go-expert/rate-limiter/pkg/repository"
	"github.com/go-chi/chi"
)

func main() {
	log.Println("Reading config files...")
	config, err := config.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	log.Println("Creating repository...")
	repo := repository.NewSQLite("configs.db")

	wsAddress := fmt.Sprintf("%s:%s", config.WebServerHost, config.WebServerPort)
	log.Println("Creating web server...")
	ws := webserver.NewWebServer(
		chi.NewRouter(),
		repo,
		wsAddress,
		config,
	)
	err = repo.Connect()
	if err != nil {
		panic(err)
	}

	log.Println("Creating cache...")
	cacheAddress := fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort)

	log.Println("Connecting to Redis cache...")
	cache, err := cache.NewRedisCache(cacheAddress, config.RedisReadTimeout, config.RedisWriteTimeout)
	if err != nil {
		panic(err)
	}

	log.Println("Creating rate limiter...")
	rateLimiterMiddleware := httprate.NewRateLimiter(repo, cache, config.RateLimitMaxRequests, config.RateLimitBlockTime)

	log.Println("Setup middleware...")
	ws.AddMiddleware("rateLimiterMiddleware", rateLimiterMiddleware.Limit)
	log.Println("Setup handlers...")
	ws.AddHandler(http.MethodGet, "/rate-limit", ws.RateLimiterHandler())
	ws.AddHandler(http.MethodGet, "/config", ws.GetConfigByID())
	ws.AddHandler(http.MethodGet, "/configs", ws.GetAllConfigs())
	ws.AddHandler(http.MethodPost, "/config", ws.CreateConfig())
	ws.AddHandler(http.MethodPatch, "/config", ws.UpdateConfig())
	ws.AddHandler(http.MethodDelete, "/config", ws.DeleteConfig())

	log.Println("Starting web server...")
	ws.Start()
}
