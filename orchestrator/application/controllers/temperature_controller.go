package controllers

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
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

func (c *TemperatureController) GetTemperature(ctx *fiber.Ctx) error {
	log.Printf("Iniciando processamento Controller")

	cep := ctx.Query("cep")
	// clientContext := ctx.Query("ctx")

	log.Printf("ctx: %v", ctx.Query("ctx"))

	var context context.Context
	ctx.QueryParser(&context)

	log.Printf("CONTEXT: %v", context)

	_, spanInicial := c.otelTracer.Start(context, "Controller-GetTemperature-Span")
	defer spanInicial.End()

	if cep == "" || len(cep) != 8 {
		return ctx.Status(422).JSON(response.ErrorResponse{
			Error: "invalid zipcode",
		})
	}

	temperatures, err := c.temperatureUseCase.GetTemperature(context, cep)
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
