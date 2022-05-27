package cloudeventhttpexporter

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"time"
)

import (
	"context"
	"errors"
)

const (
	// The value of "type" key in configuration.
	typeStr = "cloudevent"

	defaultTimeout = time.Second * 5

	defaultFormat = "json"
)

// NewFactory creates a factory for Cloud Event exporter.
func NewFactory() component.ExporterFactory {
	return component.NewExporterFactory(
		typeStr,
		createDefaultConfig,
		component.WithTracesExporter(createTracesExporter))
}

func createDefaultConfig() config.Exporter {
	return &Config{
		ExporterSettings: config.NewExporterSettings(config.NewComponentID(typeStr)),
		RetrySettings:    exporterhelper.NewDefaultRetrySettings(),
		QueueSettings:    exporterhelper.NewDefaultQueueSettings(),
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Timeout: defaultTimeout,
			// We almost read 0 bytes, so no need to tune ReadBufferSize.
			WriteBufferSize: 512 * 1024,
		},
		Format: defaultFormat,
	}
}

func createTracesExporter(
	_ context.Context,
	set component.ExporterCreateSettings,
	cfg config.Exporter,
) (component.TracesExporter, error) {
	zc := cfg.(*Config)

	if zc.Endpoint == "" {
		// TODO https://github.com/open-telemetry/opentelemetry-collector/issues/215
		return nil, errors.New("exporter config requires a non-empty 'endpoint'")
	}

	ze, err := createCloudEventExporter(zc, set.TelemetrySettings)
	if err != nil {
		return nil, err
	}
	return exporterhelper.NewTracesExporter(
		zc,
		set,
		ze.pushTraces,
		exporterhelper.WithTimeout(exporterhelper.TimeoutSettings{Timeout: 0}),
		exporterhelper.WithQueue(zc.QueueSettings),
		exporterhelper.WithRetry(zc.RetrySettings))
}
