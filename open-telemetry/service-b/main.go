package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	tracer  trace.Tracer
	Current struct {
		TempC tempC `json:"temp_c"`
		TempF tempF `json:"temp_f"`
		TempK tempK `json:"temp_k"`
	} `json:"current"`
}

func (w *weatherAPI) setTracer(name string) {
	w.tracer = otel.Tracer(name)
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

func (w *weatherAPI) getWeather(r *http.Request, localidade string) (*weatherAPI, error) {
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	ctx, span := w.tracer.Start(ctx, "request-to-service-weatherAPI")
	defer span.End()
	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", w.apiKey, url.QueryEscape(localidade))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	otel.GetTextMapPropagator().Inject(ctx, carrier)
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
	Localidade string `json:"city"`
	TempC      tempC  `json:"temp_C"`
	TempF      tempF  `json:"temp_F"`
	TempK      tempK  `json:"temp_K"`
}

func (r *response) setLocalidade(localidade string) {
	r.Localidade = localidade

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
	tracer     trace.Tracer
}

func (v *viaCEP) getLocalidade(r *http.Request, cep string) string {
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	ctx, span := v.tracer.Start(ctx, "request-to-service-viaCEP")
	defer span.End()
	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ""
	}

	req.Header.Set("Accept", "application/json")

	otel.GetTextMapPropagator().Inject(ctx, carrier)
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

	bsp := tracesdk.NewBatchSpanProcessor(tracerExporter)
	traceProvider := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithResource(res),
		tracesdk.WithSpanProcessor(bsp),
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

	shutdown, err := initOtelProvider(ctx, "service-b", "collector:4317")
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := shutdown(ctx); err != nil {
			panic(err)
		}
	}()
	apiKey := os.Getenv("WEATHER_API_KEY")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := new(http.Client)
	client.Transport = tr

	defer client.CloseIdleConnections()

	tracer := otel.Tracer("service-b-tracer")
	viaCEPTracer := otel.Tracer("service-b-viaCEP-tracer")
	weatherAPITracer := otel.Tracer("service-b-weatherAPI-tracer")

	viaCEP := &viaCEP{client: client, tracer: viaCEPTracer}
	weatherAPI := &weatherAPI{apiKey: apiKey, client: client, tracer: weatherAPITracer}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(time.Second * 60))

	router.Get("/weather", func(w http.ResponseWriter, r *http.Request) {
		carrier := propagation.HeaderCarrier(r.Header)
		ctx := r.Context()
		ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
		ctx, span := tracer.Start(ctx, "request-from-service-b-to-external-services")
		otel.GetTextMapPropagator().Inject(ctx, carrier)
		defer span.End()
		cep := r.URL.Query().Get("cep")
		if strings.Count(cep, "")-1 != cepLength {
			http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
			return
		}

		localidade := viaCEP.getLocalidade(r, cep)
		log.Println("localidade: ", localidade)
		if localidade == "" {
			http.Error(w, "can not find zipcode", http.StatusNotFound)
			return
		}

		weather, err := weatherAPI.getWeather(r, localidade)
		log.Println("weather: ", weather)
		if err != nil {
			fmt.Println(err)
			return
		}

		response := new(response)
		response.setLocalidade(localidade)
		response.setTempC(weather.tempC())
		response.setTempF(weather.tempCToF())
		response.setTempK(weather.tempCToK())

		log.Println("response from /weather")
		log.Println(response)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	http.ListenAndServe(":8080", router)
}
