package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/murilocarbol/observability-and-telemetry/application/client/response"
	"github.com/murilocarbol/observability-and-telemetry/application/model"
	tracer "go.opentelemetry.io/otel/trace"
)

type WeatherClient struct {
	api_key    string
	otelTracer tracer.Tracer
}

func NewWeatherClient(key string, trace tracer.Tracer) *WeatherClient {
	return &WeatherClient{
		api_key:    key,
		otelTracer: trace,
	}
}

type WeatherClientInterface interface {
	GetWeather(localitation string) (*model.Temperature, error)
}

func (v WeatherClient) GetWeather(ctx context.Context, localitation string) (*model.Temperature, error) {
	_, span := v.otelTracer.Start(ctx, "Client-GetWeather-Span")
	defer span.End()

	req, err := http.NewRequest("GET", "http://api.weatherapi.com/v1/current.json", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("key", v.api_key)
	q.Add("q", localitation)
	q.Add("aqi", "no")
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var weather response.Weather
	err = json.Unmarshal(body, &weather)
	if err != nil {
		return nil, err
	}

	temperature := &model.Temperature{
		TemperatureCelsius:   weather.Current.TempC,
		TemperatureFarenheit: weather.Current.TempF,
	}

	return temperature, nil
}
