package metricscollector

import (
	"fmt"
	"net/http"

	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func NewMeterProvider() (metric.MeterProvider, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("failed to setup exporter for meter provider: %w", err)
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("kessel-inventory-consumer"),
			),
		),
		sdkmetric.WithReader(exporter),
		sdkmetric.WithView(
			metrics.DefaultSecondsHistogramView(metrics.DefaultServerSecondsHistogramName),
		),
	)
	otel.SetMeterProvider(provider)
	return provider, nil
}

func NewMeter(provider metric.MeterProvider) (metric.Meter, error) {
	return provider.Meter("kessel-inventory-consumer"), nil
}

func ServeMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":9000", nil)
	if err != nil {
		fmt.Printf("error serving metrics: %v", err)
		return
	}
}
