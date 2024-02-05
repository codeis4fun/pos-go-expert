package httprate

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/codeis4fun/pos-go-expert/rate-limiter/pkg/cache"
	"github.com/codeis4fun/pos-go-expert/rate-limiter/pkg/repository"
)

type RateLimiter struct {
	repository              repository.Repository
	cache                   cache.Cache
	requestLimit, blockTime int
}

func NewRateLimiter(repository repository.Repository, cache cache.Cache, requestLimit, blockTime int) *RateLimiter {
	return &RateLimiter{
		repository:   repository,
		cache:        cache,
		requestLimit: requestLimit,
		blockTime:    blockTime,
	}
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.RemoteAddr
		if r.Header.Get("API_KEY") != "" {
			key = r.Header.Get("API_KEY")
		}

		config, err := rl.repository.GetConfigByConfigValue(key)
		if err != nil && err.Error() != "record not found" {
			log.Println("failed to get config by id: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if config.MaxRequest == 0 {
			config.MaxRequest = rl.requestLimit
			log.Println("setting request limit to:", config.MaxRequest)
		}
		if config.BlockTime == 0 {
			config.BlockTime = rl.blockTime
			log.Println("setting block time to:", config.BlockTime)
		}

		val, err := rl.cache.Get(key)
		if err != nil {
			log.Println("failed to get value from cache: ", err)
			return
		}

		if val == "" {
			val = "1"
			err := rl.cache.Set(key, val, time.Duration(config.BlockTime)*time.Second)
			if err != nil {
				log.Println("failed to set value in cache: ", err)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		count, err := strconv.Atoi(val)
		if err != nil {
			log.Println("failed to convert value to int: ", err)
			return
		}

		if count+1 > config.MaxRequest {
			log.Println("too many requests")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("you have reached the maximum number of requests or actions allowed within a certain time frame"))
			return
		}

		err = rl.cache.Set(key, strconv.Itoa(count+1), time.Duration(config.BlockTime)*time.Second)
		if err != nil {
			log.Println("failed to set value in cache: ", err)
			return
		}

		next.ServeHTTP(w, r)
	})
}
