package cloudeventhttpexporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"net/http"
)

type cloudEventExporter struct {
	url   string
	token string

	client *http.Client

	clientSettings *confighttp.HTTPClientSettings
	settings       component.TelemetrySettings
}

func createCloudEventExporter(cfg *Config, settings component.TelemetrySettings) (*cloudEventExporter, error) {
	exporter := &cloudEventExporter{
		url:   cfg.Endpoint,
		token: cfg.Format,

		client: nil,

		clientSettings: &cfg.HTTPClientSettings,
		settings:       settings,
	}

	fmt.Println("Creating CloudEvent exporter")

	return exporter, nil
}

// start creates the http client
func (ce *cloudEventExporter) start(_ context.Context, host component.Host) (err error) {
	ce.client, err = ce.clientSettings.ToClient(host.GetExtensions(), ce.settings)
	return
}

func (ce *cloudEventExporter) pushTraces(ctx context.Context, td ptrace.Traces) error {
	fmt.Println("Got a Traces")
	var batch []cloudevents.Event

	for i := 0; i < 10; i++ {
		event := ce.createEvent(i)
		batch = append(batch, event)
	}

	err := ce.buildAndSendBatch(ctx, batch)
	if err != nil {
		return err
	}

	return nil
}

func (ce *cloudEventExporter) createEvent(id int) cloudevents.Event {
	fmt.Println("Creating a cloud event")
	event := cloudevents.NewEvent()

	event.SetSpecVersion("1.0")
	event.SetID(uuid.New().String())
	event.SetType("com.cloudevents.sample.sent")
	event.SetExtension("traceid", "test")
	event.SetExtension("group", "otel")
	event.SetSource("https://github.com/cloudevents/sdk-go/v2/samples/httpb/sender")
	_ = event.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"id":      id,
		"message": "Hello, World!",
	})
	return event
}

func (ce *cloudEventExporter) buildAndSendBatch(ctx context.Context, batch []cloudevents.Event) error {
	fmt.Println("Sending cloud events (token: " + ce.token + ")")

	j, _ := json.Marshal(batch)
	fmt.Println(string(j))

	req, err := http.NewRequestWithContext(ctx, "POST", ce.url, bytes.NewReader(j))
	if err != nil {
		return fmt.Errorf("failed to push trace data via Zipkin exporter: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+ce.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ce.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send cloud event: %w", err)
	}
	fmt.Println("Sent cloud events with code: " + resp.Status)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("failed the request with status code %s", resp.Status)
	}
	return nil
}
