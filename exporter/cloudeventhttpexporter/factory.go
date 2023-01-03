package cloudeventhttpexporter

import (
	"context"
	"errors"

	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

const (
	// The value of "type" key in configuration.
	typeStr = "cloudevent"

	defaultTimeout = time.Second * 5

	defaultFormat = "json"
	// The stability level of the exporter.
	stability = component.StabilityLevelBeta
)

// NewFactory creates a factory for Cloud Event exporter.
func NewFactory() exporter.Factory {
	return component.NewExporterFactory(
		typeStr,
		createDefaultConfig,
		component.WithTracesExporter(createTracesExporter, stability))
}

func createDefaultConfig() component.Config {
	return &Config{
		ExporterSettings: config.NewExporterSettings(component.NewID(typeStr)),
		RetrySettings:    exporterhelper.NewDefaultRetrySettings(),
		QueueSettings:    exporterhelper.NewDefaultQueueSettings(),
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Timeout: defaultTimeout,
			// We almost read 0 bytes, so no need to tune ReadBufferSize.
			WriteBufferSize: 512 * 1024,
		},
		AccessToken: "",
	}
}

func createTracesExporter(
	ctx context.Context,
	set component.ExporterCreateSettings,
	cfg component.Config,
) (component.TracesExporter, error) {
	cc := cfg.(*Config)

	if cc.Endpoint == "" {
		// TODO https://github.com/open-telemetry/opentelemetry-collector/issues/215
		return nil, errors.New("exporter config requires a non-empty 'endpoint'")
	}

	ce, err := createCloudEventExporter(cc, set.TelemetrySettings)
	if err != nil {
		return nil, err
	}
	return exporterhelper.NewTracesExporter(
		ctx,
		set,
		cfg,
		ce.pushTraces,
		exporterhelper.WithStart(ce.start),
		// explicitly disable since we rely on http.Client timeout logic.
		exporterhelper.WithTimeout(exporterhelper.TimeoutSettings{Timeout: 0}),
		exporterhelper.WithQueue(cc.QueueSettings),
		exporterhelper.WithRetry(cc.RetrySettings))
}
