package presidioprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

var (
	// TypeStr is the identifier for your processor in the config.yaml
	TypeStr = component.MustNewType("presidio")
)

// NewFactory creates a factory for the custom processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		TypeStr,
		createDefaultConfig,
		processor.WithTraces(createTracesProcessor, component.StabilityLevelDevelopment),
		processor.WithLogs(createLogsProcessor, component.StabilityLevelDevelopment),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

// Inside factory.go
func createTracesProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Traces,
) (processor.Traces, error) {

	oCfg := cfg.(*Config)

	// Pass the configured client and the config to your processor
	return newTracesProcessor(set.Logger, set.TelemetrySettings, oCfg, nextConsumer)
}

// Inside factory.go
func createLogsProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (processor.Logs, error) {

	oCfg := cfg.(*Config)

	// Pass the configured client and the config to your processor
	return newLogsProcessor(set.Logger, set.TelemetrySettings, oCfg, nextConsumer)
}
