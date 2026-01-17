package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// Order
type Order struct {
	ID     string  `json:"id"`
	Amount float64 `json:"amount"`
}

var tracer trace.Tracer

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

func tracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем контекст из заголовков HTTP запроса
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		// Создаем span и продолжаем обработку
		ctx, span := tracer.Start(ctx, "http_request")
		defer span.End()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	// Инициализация трейсинга
	tp, err := initTracer("service-a")
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("Error shutting down tracer provider: %v", err)
		}
	}()

	http.Handle("/api/calc", tracingMiddleware(http.HandlerFunc(handleOrders)))
	http.Handle("/api/health", tracingMiddleware(http.HandlerFunc(handleHealth)))

	port := "8081"

	log.Printf("Starting service on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, span := tracer.Start(ctx, "health_check")
	defer span.End()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}

func handleOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	defer span.End()

	// Добавляем атрибуты к спану
	span.SetAttributes(
		semconv.HTTPMethodKey.String(r.Method),
		semconv.HTTPURLKey.String(r.URL.String()),
	)

	switch r.Method {
	case "GET":
		id := r.URL.Query().Get("id")
		if id != "" {
			calcOrder(w, r, id)
		} else {
			calcOrders(w, r)
		}
	default:
		span.SetAttributes(semconv.HTTPStatusCodeKey.Int(http.StatusMethodNotAllowed))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func calcOrder(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	defer span.End()

	// Добавляем атрибуты заказа к трассировке
	span.SetAttributes(
		semconv.HTTPStatusCodeKey.Int(http.StatusOK),
	)

	// Имитация обработки заказа
	processOrder(ctx, id)

	order := Order{
		ID:     id,
		Amount: 100,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func calcOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.SetAttributes(semconv.HTTPStatusCodeKey.Int(http.StatusOK))

	// Имитация обработки нескольких заказов
	processMultipleOrders(ctx)

	orders := []Order{
		{ID: "1", Amount: 100},
		{ID: "2", Amount: 200},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func processOrder(ctx context.Context, orderID string) {
	// Создаем дочерний спан для обработки заказа
	ctx, span := tracer.Start(ctx, "process_order")
	defer span.End()

	// Добавляем кастомные атрибуты
	span.SetAttributes(
		attribute.String("order.id", orderID),
		attribute.String("operation.type", "process"),
		attribute.String("operation.name", "order_processing"),
	)

	// Имитация обработки
	time.Sleep(10 * time.Millisecond)
	span.AddEvent("order_processed")
}

func processMultipleOrders(ctx context.Context) {
	// Создаем дочерний спан для обработки нескольких заказов
	ctx, span := tracer.Start(ctx, "process_multiple_orders")
	defer span.End()

	// Добавляем кастомные атрибуты
	span.SetAttributes(
		attribute.String("operation.type", "batch_process"),
		attribute.String("operation.name", "bulk_order_processing"),
		attribute.Int("order.count", 2),
	)

	// Имитация обработки
	time.Sleep(20 * time.Millisecond)
	span.AddEvent("multiple_orders_processed")
}
