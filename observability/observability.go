package observability

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Init sets up a zap logger and an OpenTelemetry tracer provider. It returns
// (logger, tracerProvider, shutdownFunc, error). Shutdown may be nil if tracer
// couldn't be created.
func Init(serviceName string) (*zap.Logger, *sdktrace.TracerProvider, func(context.Context) error, error) {
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "json"
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := cfg.Build()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to build logger: %w", err)
	}

	// Create OTLP exporter (HTTP) for local/dev. If it fails, return logger
	// and nil tracer provider so services can still log.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "otel-collector:4318"
	}

	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		logger.Warn("failed to create OTLP exporter, tracing disabled", zap.Error(err))
		return logger, nil, func(ctx context.Context) error { return nil }, nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceNameKey.String(serviceName)),
	)
	if err != nil {
		logger.Warn("failed to create resource for tracer", zap.Error(err))
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	shutdown := func(ctx context.Context) error {
		_ = logger.Sync()
		return tp.Shutdown(ctx)
	}

	return logger, tp, shutdown, nil
}
