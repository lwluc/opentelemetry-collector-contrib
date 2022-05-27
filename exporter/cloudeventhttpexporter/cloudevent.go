package cloudeventhttpexporter

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"log"
)

type cloudEventExporter struct {
	url string
}

func createCloudEventExporter(cfg *Config, settings component.TelemetrySettings) (*cloudEventExporter, error) {
	exporter := &cloudEventExporter{
		url: cfg.Endpoint,
	}

	return exporter, nil
}

func (ce *cloudEventExporter) pushTraces(ctx context.Context, td ptrace.Traces) error {
	ctxCloudEvent := cloudevents.ContextWithTarget(context.Background(), ce.url)

	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	client, err := cloudevents.NewClient(p, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	for i := 0; i < 10; i++ {
		e := cloudevents.NewEvent()

		e.SetSpecVersion("1.0")
		e.SetID(string(i))
		e.SetType("com.cloudevents.sample.sent")
		err := e.ExtensionAs("traceid", "test")
		if err != nil {
			return err
		}
		err2 := e.ExtensionAs("group", "otel")
		if err2 != nil {
			return err2
		}
		e.SetSource("https://github.com/cloudevents/sdk-go/v2/samples/httpb/sender")
		_ = e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
			"id":      i,
			"message": "Hello, World!",
		})

		res := client.Send(ctxCloudEvent, e)
		if cloudevents.IsUndelivered(res) {
			log.Printf("Failed to send: %v", res)
		} else {
			var httpResult *cehttp.Result
			cloudevents.ResultAs(res, &httpResult)
			log.Printf("Sent %d with status code %d", i, httpResult.StatusCode)
		}
	}

	return nil
}
