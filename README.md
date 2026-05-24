# Presidio OpenTelemetry Processor

This otel processor is used to detect and redact sensitive data in logs and traces send via otlp through a otel collector pipeline. It is based on Microsoft's Presidio project.

## Build a custom otel collector with this processor

Use [OCB](https://opentelemetry.io/docs/collector/extend/ocb/) to build a custom otel collector with this processor. Below is an example of the OCB configuration file to build a custom otel collector with this processor.

```yml
dist:
  name: otelcol-dev
  description: Basic OTel Collector distribution for Developers
  output_path: ./otelcol-dev

exporters:
  - gomod:
      go.opentelemetry.io/collector/exporter/debugexporter v0.152.0
  - gomod:
      go.opentelemetry.io/collector/exporter/otlpexporter v0.152.0

processors:
  - gomod: go.opentelemetry.io/collector/processor/batchprocessor v0.152.0
  - gomod: github.com/mxab/otel-presidio/processor v0.0.0
    name: presido

 

receivers:
  - gomod:
      go.opentelemetry.io/collector/receiver/otlpreceiver v0.152.0

providers:
  - gomod:
      go.opentelemetry.io/collector/confmap/provider/envprovider v1.48.0
  - gomod:
      go.opentelemetry.io/collector/confmap/provider/fileprovider v1.48.0
  - gomod:
      go.opentelemetry.io/collector/confmap/provider/httpprovider v1.48.0
  - gomod:
      go.opentelemetry.io/collector/confmap/provider/httpsprovider v1.48.0
  - gomod:
      go.opentelemetry.io/collector/confmap/provider/yamlprovider v1.48.0

```

## Thanks to this blog for the instructions on how to build a custom otel collector processor:

https://oneuptime.com/blog/post/2026-02-06-otel-custom-collector-processor/view