package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type CalcService struct {
	baseURL string
	client  *http.Client
	tracer  trace.Tracer
}

func NewCalcService(baseURL string) *CalcService {
	return &CalcService{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
		tracer:  otel.Tracer("calc-service-client"),
	}
}

func (s *CalcService) CalcOrder(ctx context.Context, id string) error {
	// Создаем дочерний спан для вызова сервиса A
	ctx, span := s.tracer.Start(ctx, "call_service_a")
	defer span.End()

	// Создаем запрос с контекстом трассировки
	url := fmt.Sprintf("%s/api/calc?id=%s", s.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("creating request: %w", err)
	}

	// Инжектируем заголовки трассировки в запрос
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Выполняем запрос
	resp, err := s.client.Do(req)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("calling service A: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("service A returned status %d: %s", resp.StatusCode, string(body))
		span.RecordError(err)
		return err
	}

	return nil
}
