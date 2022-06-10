package cloudeventhttpexporter

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
	"net/http"
)

type cloudEventExporter struct {
	url   string
	token string

	client *http.Client

	logger *zap.Logger
}

func createCloudEventExporter(cfg *Config, settings component.TelemetrySettings) (*cloudEventExporter, error) {
	exporter := &cloudEventExporter{
		url:   cfg.Endpoint,
		token: cfg.Format,
	}

	fmt.Println("Creating CloudEvent exporter")

	return exporter, nil
}

func (ce *cloudEventExporter) pushTraces(ctx context.Context, td ptrace.Traces) error {
	ce.logger.Info("Got a Traces")
	var batch []cloudevents.Event

	for i := 0; i < 10; i++ {
		e, err := ce.createEvent(i)
		if err != nil {
			return err
		}

		batch = append(batch, e)
	}

	err2 := ce.buildAndSendBatch(ctx, batch)
	if err2 != nil {
		return err2
	}

	return nil
}

func (ce *cloudEventExporter) createEvent(i int) (cloudevents.Event, error) {
	ce.logger.Info("Creating a cloud event")
	event := cloudevents.NewEvent()

	event.SetSpecVersion("1.0")
	event.SetID(string(i))
	event.SetType("com.cloudevents.sample.sent")
	err := event.ExtensionAs("traceid", "test")
	if err != nil {
		return event, err
	}
	err2 := event.ExtensionAs("group", "otel")
	if err2 != nil {
		return event, err2
	}
	event.SetSource("https://github.com/cloudevents/sdk-go/v2/samples/httpb/sender")
	_ = event.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"id":      i,
		"message": "Hello, World!",
	})
	return event, err
}

func (ce *cloudEventExporter) buildAndSendBatch(ctx context.Context, batch []cloudevents.Event) error {
	ce.logger.Info("Setting up own telemetry...")
	ce.logger.Info("Sending cloud events")
	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(batch)
	body := buf.Bytes()

	req, err := http.NewRequestWithContext(ctx, "POST", ce.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to push trace data via Zipkin exporter: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+ce.token)
	req.Header.Set("Content-Length", string(binary.Size(body)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := ce.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send cloud event: %w", err)
	}
	ce.logger.Info("Sent cloud events with code: " + string(resp.StatusCode))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("failed the request with status code %d", resp.StatusCode)
	}
	return nil
}
