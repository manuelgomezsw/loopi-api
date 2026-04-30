package observability

import (
	"context"
	"fmt"

	mexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// InitMeter sets up the OpenTelemetry MeterProvider with Google Cloud Monitoring as exporter.
// Metrics are exported every 60s (Cloud Monitoring minimum interval).
// When projectID is empty (local dev), the global provider stays as no-op.
// Returns a shutdown function that must be deferred by the caller.
func InitMeter(projectID string) (func(context.Context) error, error) {
	noop := func(context.Context) error { return nil }

	if projectID == "" {
		return noop, nil
	}

	exporter, err := mexporter.New(
		mexporter.WithProjectID(projectID),
	)
	if err != nil {
		return nil, fmt.Errorf("cloud monitoring exporter: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			// Cloud Monitoring minimum export interval is 60s
			sdkmetric.NewPeriodicReader(exporter,
				sdkmetric.WithInterval(60_000_000_000), // 60s in nanoseconds
			),
		),
	)

	otel.SetMeterProvider(mp)

	return mp.Shutdown, nil
}
