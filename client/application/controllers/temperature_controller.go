package controllers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/murilocarbol/observability-and-telemetry/application/controllers/request"
	"github.com/murilocarbol/observability-and-telemetry/application/controllers/response"
	"github.com/murilocarbol/observability-and-telemetry/application/usecases"
	tracer "go.opentelemetry.io/otel/trace"
)

type TemperatureController struct {
	temperatureUseCase usecases.TemperatureUseCase
	otelTracer         tracer.Tracer
}

func NewTemperatureController(temperatureUsecase *usecases.TemperatureUseCase, trace tracer.Tracer) *TemperatureController {
	return &TemperatureController{
		temperatureUseCase: *temperatureUsecase,
		otelTracer:         trace,
	}
}

func (c *TemperatureController) PostTemperature(ctx *fiber.Ctx) error {
	log.Printf("Iniciando processamento Controller")

	context, spanInicial := c.otelTracer.Start(ctx.Context(), "Controller-PostTemperature-Span")
	defer spanInicial.End()

	// _, span := c.otelTracer.Start(context, "Teste-Span")
	// defer span.End()

	var request = request.TemperatureRequest{}
	err := ctx.BodyParser(&request)
	if err != nil {
		return ctx.Status(400).JSON(response.ErrorResponse{Error: err.Error()})
	}

	log.Printf("request %v", request)

	temperatures, err := c.temperatureUseCase.GetTemperature(context, request.Cep)
	if err != nil {
		if err.Error() == "zipcode not found" {
			return ctx.Status(404).JSON(response.ErrorResponse{Error: err.Error()})
		}
		return ctx.Status(500).JSON(response.ErrorResponse{
			Error: "internal server error",
		})
	}

	log.Printf("Temperatures: %+v", temperatures)

	temperaturesResponse := &response.TemperatureResponse{
		City:                 temperatures.City,
		TemperatureCelsius:   temperatures.TemperatureCelsius,
		TemperatureFarenheit: temperatures.TemperatureFarenheit,
		TemperatureKelvin:    temperatures.TemperatureKelvin,
	}

	return ctx.Status(200).JSON(temperaturesResponse)
}
