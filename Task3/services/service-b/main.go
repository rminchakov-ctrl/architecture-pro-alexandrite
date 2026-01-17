package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/architecture-pro-alexandrite/task3/services/service-b/services"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

var calcService *services.CalcService
var tracer trace.Tracer

// Order
type Order struct {
	ID     string  `json:"id"`
	Amount float64 `json:"amount"`
}

func initTracer(srvName string) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint("simplest-collector.observability.svc.cluster.local:4318"),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Создаем ресурс с информацией о сервисе
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(srvName),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Создаем провайдер трассировки
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Устанавливаем глобальный провайдер трассировки
	otel.SetTracerProvider(tp)

	// Устанавливаем пропагатор для распространения контекста
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tracer = tp.Tracer(srvName)
	return tp, nil
}

func main() {
	tp, err := initTracer("service-b")
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("Error shutting down tracer provider: %v", err)
		}
	}()

	caclAPIURL := "http://service-a:8081"
	calcService = services.NewCalcService(caclAPIURL)

	http.HandleFunc("/api/health", handleHealth)
	http.HandleFunc("/api/order", handleOrders)

	port := "8000"
	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, span := tracer.Start(ctx, "health_check")
	defer span.End()

	log.Print("Req health_check")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}

func handleOrders(w http.ResponseWriter, r *http.Request) {
	log.Printf("Req order start, method: %s, url:%s", r.Method, r.URL.String())

	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.SetAttributes(
		semconv.HTTPMethodKey.String(r.Method),
		semconv.HTTPURLKey.String(r.URL.String()),
	)

	switch r.Method {
	case "GET":
		id := r.URL.Query().Get("id")
		log.Printf("Req order, id=%s", id)

		if id == "" {
			http.Error(w, "ID is required", http.StatusBadRequest)
			span.SetAttributes(semconv.HTTPStatusCodeKey.Int(http.StatusBadRequest))
			return
		}

		span.SetAttributes(
			semconv.HTTPStatusCodeKey.Int(http.StatusOK),
			semconv.EnduserIDKey.String(id),
		)

		log.Print("Try to calc")
		if err := calcService.CalcOrder(ctx, id); err != nil {
			log.Printf("Error calling calc service: %v", err)
			span.SetAttributes(semconv.HTTPStatusCodeKey.Int(http.StatusInternalServerError))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		order := Order{
			ID:     id,
			Amount: 100,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	default:
		span.SetAttributes(semconv.HTTPStatusCodeKey.Int(http.StatusMethodNotAllowed))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
