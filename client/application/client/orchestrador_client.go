package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/murilocarbol/observability-and-telemetry/application/client/response"
	"github.com/murilocarbol/observability-and-telemetry/application/model"
	tracer "go.opentelemetry.io/otel/trace"
)

type OrchestradorClient struct {
	url        string
	otelTracer tracer.Tracer
}

func NewOrchestratorClient(url string, trace tracer.Tracer) *OrchestradorClient {
	return &OrchestradorClient{
		url:        url,
		otelTracer: trace,
	}
}

type OrchestradorClientInterface interface {
	CallOrchestrador(localitation string) (*model.Temperature, error)
}

func (v OrchestradorClient) CallOrchestrador(ctx context.Context, cep string) (*model.Temperature, error) {
	_, span := v.otelTracer.Start(ctx, "Client-CallOrchestrador-Span")
	defer span.End()

	req, err := http.NewRequest("GET", v.url, nil)
	if err != nil {
		return nil, err
	}

	log.Printf("Req %v", req)

	q := req.URL.Query()
	q.Add("cep", cep)
	q.Add("ctx", fmt.Sprintf("%v", ctx))
	req.URL.RawQuery = q.Encode()

	log.Printf("Req %v", req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Erro ao realizar requisição %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("Resp client %v", resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response response.OrchestradorResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	temperature := &model.Temperature{
		City:                 response.City,
		TemperatureCelsius:   response.TemperatureCelsius,
		TemperatureFarenheit: response.TemperatureFarenheit,
		TemperatureKelvin:    response.TemperatureKelvin,
	}

	return temperature, nil
}
