package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

const cepLength = 8
const serviceBEndpoint = "http://service-b:8080/weather"

type (
	temperature float64
	tempC       temperature
	tempF       temperature
	tempK       temperature
)

type response struct {
	Localidade string `json:"city"`
	TempC      tempC  `json:"temp_C"`
	TempF      tempF  `json:"temp_F"`
	TempK      tempK  `json:"temp_K"`
}

func initOtelProvider(ctx context.Context, serviceName, serviceURL string) (shutdown func(context.Context) error, err error) {
	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(serviceName)))
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		serviceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	tracerExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	bsp := trace.NewBatchSpanProcessor(tracerExporter)
	traceProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(traceProvider)

	otel.SetTextMapPropagator(propagation.TraceContext{})

	return traceProvider.Shutdown, nil
}

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	shutdown, err := initOtelProvider(ctx, "service-a", "collector:4317")
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := shutdown(ctx); err != nil {
			panic(err)
		}
	}()

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(time.Second * 60))

	tracer := otel.Tracer("service-a-tracer")

	router.Get("/weather/{cep}", func(w http.ResponseWriter, r *http.Request) {
		carrier := propagation.HeaderCarrier(r.Header)
		ctx := r.Context()
		ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
		ctx, span := tracer.Start(ctx, "request-from-service-a-to-service-b")
		defer span.End()
		cep := chi.URLParam(r, "cep")
		if len(cep) != cepLength {
			http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
			return
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s?cep=%s", serviceBEndpoint, cep), nil)

		if err != nil {
			http.Error(w, "service b error", http.StatusInternalServerError)
			return
		}

		otel.GetTextMapPropagator().Inject(ctx, carrier)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "service b error", http.StatusInternalServerError)
			return
		}

		defer resp.Body.Close()

		var response response

		json.NewDecoder(resp.Body).Decode(&response)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		json.NewEncoder(w).Encode(response)
	})

	http.ListenAndServe(":8090", router)

}
