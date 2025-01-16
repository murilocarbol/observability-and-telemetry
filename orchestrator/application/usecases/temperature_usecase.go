package usecases

import (
	"context"
	"fmt"
	"log"

	"github.com/murilocarbol/observability-and-telemetry/application/client"
	"github.com/murilocarbol/observability-and-telemetry/application/model"
	tracer "go.opentelemetry.io/otel/trace"
)

type TemperatureUseCaseInterface interface {
	GetTemperature(cep string) (*model.Temperature, error)
}

type TemperatureUseCase struct {
	viaCepClient  client.ViaCepClient
	weatherClient client.WeatherClient
	otelTracer    tracer.Tracer
}

func NewTemperatureUseCase(viaCepClient *client.ViaCepClient, weatherClient *client.WeatherClient, trace tracer.Tracer) *TemperatureUseCase {
	return &TemperatureUseCase{
		viaCepClient:  *viaCepClient,
		weatherClient: *weatherClient,
		otelTracer:    trace,
	}
}

func (t *TemperatureUseCase) GetTemperature(ctx context.Context, cep string) (*model.Temperature, error) {
	_, span := t.otelTracer.Start(ctx, "UseCase-GetTemperature-Span")
	defer span.End()

	city, err := t.viaCepClient.GetEndereco(ctx, cep)
	if err != nil {
		return nil, fmt.Errorf("zipcode not found")
	}

	log.Printf("City: %s", city)

	temp, err := t.weatherClient.GetWeather(ctx, city)
	if err != nil {
		return nil, err
	}

	temp.City = city
	temp.TemperatureKelvin = temp.TemperatureCelsius + 273

	return temp, nil
}
