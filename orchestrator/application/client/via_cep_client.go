package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/murilocarbol/observability-and-telemetry/application/client/response"
	tracer "go.opentelemetry.io/otel/trace"
)

type ViaCepClient struct {
	otelTracer tracer.Tracer
}

func NewViaCepClient(trace tracer.Tracer) *ViaCepClient {
	return &ViaCepClient{
		otelTracer: trace,
	}
}

type ViaCepClientInterface interface {
	GetEndereco(cep string) (string, error)
}

func (v ViaCepClient) GetEndereco(ctx context.Context, cep string) (string, error) {
	_, span := v.otelTracer.Start(ctx, "Client-GetEndereco-Span")
	defer span.End()

	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var endereco response.Endereco
	err = json.Unmarshal(body, &endereco)
	if err != nil {
		return "", err
	}

	if endereco.Erro != "" {
		return "", fmt.Errorf(endereco.Erro)
	}

	return endereco.Localidade, nil
}
