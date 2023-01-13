package cloudeventhttpexporter

import (
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

// Config defines configuration settings for the Cloud Event HTTP exporter.
type Config struct {
	exporterhelper.QueueSettings `mapstructure:"sending_queue"`
	exporterhelper.RetrySettings `mapstructure:"retry_on_failure"`

	// Configures the exporter client.
	// The Endpoint to send the Cloud Events to (e.g.: http://some.url:9411/api/v2/spans).
	confighttp.HTTPClientSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct.

	// Token used for in header info
	AccessToken string `mapstructure:"access_token"`
}

var _ component.Config = (*Config)(nil)

// Validate checks if the exporter configuration is valid
func (cfg *Config) Validate() error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("`endpoint` not specified, please add it to your configuration file")
	}

	if cfg.AccessToken == "" {
		return fmt.Errorf("`accessToken` not specified, please add it to your configuration file")
	}

	return nil
}
