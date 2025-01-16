package usecases

import (
	"context"

	"github.com/murilocarbol/observability-and-telemetry/application/client"
	"github.com/murilocarbol/observability-and-telemetry/application/model"
	tracer "go.opentelemetry.io/otel/trace"
)

type TemperatureUseCaseInterface interface {
	GetTemperature(cep string) (*model.Temperature, error)
}

type TemperatureUseCase struct {
	orchestradorClient client.OrchestradorClient
	otelTracer         tracer.Tracer
}

func NewTemperatureUseCase(orchestradorClient *client.OrchestradorClient, trace tracer.Tracer) *TemperatureUseCase {
	return &TemperatureUseCase{
		orchestradorClient: *orchestradorClient,
		otelTracer:         trace,
	}
}

func (t *TemperatureUseCase) GetTemperature(ctx context.Context, cep string) (*model.Temperature, error) {
	_, span := t.otelTracer.Start(ctx, "UseCase-GetTemperature-Span")
	defer span.End()

	temp, err := t.orchestradorClient.CallOrchestrador(ctx, cep)
	if err != nil {
		return nil, err
	}

	return temp, nil
}
