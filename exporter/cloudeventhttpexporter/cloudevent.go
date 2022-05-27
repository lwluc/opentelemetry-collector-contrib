package cloudeventhttpexporter

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"net/http"
)

type cloudEventExporter struct {
	url   string
	token string

	client *http.Client
}

func createCloudEventExporter(cfg *Config, settings component.TelemetrySettings) (*cloudEventExporter, error) {
	exporter := &cloudEventExporter{
		url:   cfg.Endpoint,
		token: cfg.Format,
	}

	return exporter, nil
}

func (ce *cloudEventExporter) pushTraces(ctx context.Context, td ptrace.Traces) error {
	var batch []event.Event

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

func (ce *cloudEventExporter) createEvent(i int) (event.Event, error) {
	e := cloudevents.NewEvent()

	e.SetSpecVersion("1.0")
	e.SetID(string(i))
	e.SetType("com.cloudevents.sample.sent")
	err := e.ExtensionAs("traceid", "test")
	if err != nil {
		return event.Event{}, err
	}
	err2 := e.ExtensionAs("group", "otel")
	if err2 != nil {
		return event.Event{}, err2
	}
	e.SetSource("https://github.com/cloudevents/sdk-go/v2/samples/httpb/sender")
	_ = e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"id":      i,
		"message": "Hello, World!",
	})
	return e, err
}

func (ce *cloudEventExporter) buildAndSendBatch(ctx context.Context, batch []event.Event) error {
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
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("failed the request with status code %d", resp.StatusCode)
	}
	return nil
}
