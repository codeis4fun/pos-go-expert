package webserver

import (
	"log"
	"net/http"

	"github.com/codeis4fun/pos-go-expert/rate-limiter/config"
	"github.com/codeis4fun/pos-go-expert/rate-limiter/pkg/repository"
	"github.com/go-chi/chi"
)

type WebServer struct {
	Router      chi.Router
	Repository  repository.Repository
	Handlers    []handlerFunc
	Middlewares map[string]func(http.Handler) http.Handler
	Port        string
	Config      *config.Conf
}

type handlerFunc struct {
	method      string
	pattern     string
	handlerFunc http.HandlerFunc
}

func NewWebServer(
	router chi.Router,
	repository repository.Repository,
	port string,
	config *config.Conf) *WebServer {
	return &WebServer{
		Router:      router,
		Repository:  repository,
		Handlers:    []handlerFunc{},
		Middlewares: make(map[string]func(http.Handler) http.Handler),
		Port:        port,
		Config:      config,
	}
}

func (ws *WebServer) AddHandler(method, pattern string, handler http.HandlerFunc) {
	ws.Handlers = append(ws.Handlers, handlerFunc{
		method:      method,
		pattern:     pattern,
		handlerFunc: handler,
	})
}

func (ws *WebServer) AddMiddleware(name string, middleware func(http.Handler) http.Handler) {
	ws.Middlewares[name] = middleware
}

func (ws *WebServer) Start() {
	for name, middleware := range ws.Middlewares {
		log.Println("adding middleware:", name)
		ws.Router.Use(middleware)
	}
	for _, handler := range ws.Handlers {
		log.Printf("adding handler - method: %s pattern: %s handlerFunc: %v", handler.method, handler.pattern, handler.handlerFunc)
		ws.Router.MethodFunc(handler.method, handler.pattern, handler.handlerFunc)
	}
	log.Fatal(http.ListenAndServe(ws.Port, ws.Router))
}
