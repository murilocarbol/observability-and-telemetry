package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/murilocarbol/observability-and-telemetry/application/client"
	"github.com/murilocarbol/observability-and-telemetry/application/controllers"
	"github.com/murilocarbol/observability-and-telemetry/application/usecases"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type configure struct {
	TITLE                       string `mapstructure:"TITLE"`
	CONTENT                     string `mapstructure:"CONTENT"`
	EXTERNAL_CALL_URL           string `mapstructure:"EXTERNAL_CALL_URL"`
	EXTERNAL_CALL_METHOD        string `mapstructure:"EXTERNAL_CALL_METHOD"`
	REQUEST_NAME_OTEL           string `mapstructure:"REQUEST_NAME_OTEL"`
	OTEL_SERVICE_NAME           string `mapstructure:"OTEL_SERVICE_NAME"`
	OTEL_EXPORTER_OTLP_ENDPOINT string `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	HTTP_PORT                   string `mapstructure:"HTTP_PORT"`
	WEATHER_API_KEY             string `mapstructure:"WEATHER_API_KEY"`
}

func init() {
	viper.AutomaticEnv()
}

func Initialize() {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	otel_active := viper.GetString("OTEL")

	if otel_active != "false" {
		shutdown, err := initProvider(viper.GetString("OTEL_SERVICE_NAME"), viper.GetString("OTEL_EXPORTER_OTLP_ENDPOINT"))
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := shutdown(ctx); err != nil {
				log.Fatal("failed to shutdown TracerProvider: %w", err)
			}
		}()

		if err != nil {
			log.Fatal(err)
		}
	}

	confg, _ := LoadConfig(".")

	app := fiber.New()
	setRoutes(app, confg.WEATHER_API_KEY)
	app.Listen(":8181")
}

func initProvider(serviceName, collectorURL string) (func(context.Context) error, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, collectorURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider.Shutdown, nil
}

func LoadConfig(path string) (*configure, error) {
	var cfg *configure
	viper.SetConfigName("app_config")
	viper.SetConfigType("env")
	viper.AddConfigPath("./")
	viper.SetConfigFile("config.env")
	viper.AutomaticEnv()

	fmt.Println("Loading config from path:", path)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}
	return cfg, err
}

func setRoutes(app *fiber.App, key string) {

	trace := otel.Tracer("client-api")

	// Clients
	viaCepClient := client.NewViaCepClient(trace)
	weatherClient := client.NewWeatherClient(key, trace)

	// Usecases
	temperatureUseCase := usecases.NewTemperatureUseCase(viaCepClient, weatherClient, trace)

	// Controllers
	temperatureController := controllers.NewTemperatureController(temperatureUseCase, trace)

	app.Options("/", temperatureController.GetTemperature)
	app.Get("/", temperatureController.GetTemperature)
}
