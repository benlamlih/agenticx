package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/benlamlih/agenticx/internal/config"
	server "github.com/benlamlih/agenticx/internal/router"
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop() // Allow Ctrl+C to force shutdown

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	done <- true
}

func InitTracerHTTP(ctx context.Context) (*sdktrace.TracerProvider, error) {
	// Propagation (so downstream services can read our trace headers)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	cfg := config.LoadConfig()
	OtelEndpoint := cfg.App.OtelEndpoint

	endpoint := os.Getenv("OTEL_OTLP_HTTP_ENDPOINT")
	if endpoint == "" {
		endpoint = OtelEndpoint
	}

	authHeader := os.Getenv("OO_AUTH_HEADER")
	if authHeader == "" {
		authHeader = "Basic " + base64.StdEncoding.EncodeToString([]byte("admin@example.com:supersecret"))
	}

	org := os.Getenv("OO_ORG")
	if org == "" {
		org = "default"
	}

	stream := os.Getenv("OO_STREAM")
	if stream == "" {
		stream = "agenticx"
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithURLPath(fmt.Sprintf("/api/%s/v1/traces", org)),
		// TODO: Use https in production
		otlptracehttp.WithInsecure(), // use http instead of https for local dev
		otlptracehttp.WithHeaders(map[string]string{
			"authorization": authHeader,
			"organization":  org,
			"stream-name":   stream,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create otlp exporter: %w", err)
	}

	// Attach resource attributes (visible in OpenObserve UI)
	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(""),
			semconv.ServiceVersionKey.String("0.0.1"),
			attribute.String("environment", "local"),
			attribute.String("service_name", "agenticx"),
		),
	)

	// Create and register TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	return tp, nil
}

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, continuing...")
	}

	tp, err := InitTracerHTTP(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Configure structured JSON logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))
	slog.SetDefault(logger)
	httpServer := server.BuildHTTPServer(ctx, tp)

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(httpServer, done)

	slog.Info("Listening on HTTP server", "addr", httpServer.Addr)
	if err = httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
