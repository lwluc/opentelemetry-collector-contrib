# Cloud Event HTTP Exporter

<!-- TODO: Write test, clean up, ... --> 

| Status                   |                   |
| ------------------------ |-------------------|
| Stability                | [in-development]  |
| Supported pipeline types | trace             |
| Distributions            | one via this repo |

Exports data to as [CloudEvent](https://cloudevents.io) via HTTP to any back-end.
By default, this exporter uses and authentication token for POST'ing the data.

## Getting Started

The following settings are required:

- `endpoint` (no default): URL to which the exporter is going to send the CloudEvents.

By default, authentication Bearer Token is pass as HTTP-Header and must be configured under `format:`:

- `format`: Bearer Token used to authenticate against the back-end API:
  > **_NOTE:_**  This is just a hack-around to get startet. In the future there will be a separate (optional) property to set the token.

Example:

```yaml
cloudevent:
  endpoint: "http://some.location.org:9411/api/cloud-events"
  format: "access-token"
```

[in-development]:https://github.com/open-telemetry/opentelemetry-collector#in-development
