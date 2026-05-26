package presidioprocessor

import (
	"context"

	"github.com/mxab/otel-presidio/presidioprocessor/client"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

type presidioTracesProcessor struct {
	logger            *zap.Logger
	telemetrySettings component.TelemetrySettings
	config            *Config
	nextConsumer      consumer.Traces

	client client.AnonymizerServiceClient
}

func newTracesProcessor(logger *zap.Logger, telemetrySettings component.TelemetrySettings, cfg *Config, nextConsumer consumer.Traces) (*presidioTracesProcessor, error) {
	return &presidioTracesProcessor{
		logger:            logger,
		telemetrySettings: telemetrySettings,
		config:            cfg,
		nextConsumer:      nextConsumer,
	}, nil
}

func (p *presidioTracesProcessor) Start(ctx context.Context, host component.Host) error {
	p.telemetrySettings.Logger.Info("Starting presidio processor and building HTTP client")

	// 1. Extract extensions from the host safely
	var extensions map[component.ID]component.Component
	if host != nil {
		extensions = host.GetExtensions()
	}

	// 2. Use the signature you found to build the client
	clientConn, err := p.config.ClientConfig.ToClientConn(
		ctx,
		extensions,
		p.telemetrySettings,
	)
	if err != nil {
		return err // If the client fails to build, the collector will refuse to start
	}
	p.client = client.NewAnonymizerServiceClient(clientConn)

	return nil
}

func (p *presidioTracesProcessor) Shutdown(_ context.Context) error {
	p.logger.Info("Shutting down presidio traces processor")

	return nil
}

func (p *presidioTracesProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true} // Set to true since you'll likely modify traces
}

func (p *presidioTracesProcessor) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {

	traces, err := p.processTraces(ctx, td)
	if err != nil {
		return err
	}

	// 6. Pass the mutated batch to the next processor
	return p.nextConsumer.ConsumeTraces(ctx, traces)
}

// 2. The core function that OTel calls for every batch of traces
func (p *presidioTracesProcessor) processTraces(ctx context.Context, traces ptrace.Traces) (ptrace.Traces, error) {
	var textsToAnonymize []string

	// We need to keep track of which attribute corresponds to which index in our batch array
	type pointer struct {
		updateFunc func(string)
	}
	var pointers []pointer

	// Step A: Iterate through all spans and collect the target attributes
	for i := 0; i < traces.ResourceSpans().Len(); i++ {
		rs := traces.ResourceSpans().At(i)
		for j := 0; j < rs.ScopeSpans().Len(); j++ {
			ss := rs.ScopeSpans().At(j)
			for k := 0; k < ss.Spans().Len(); k++ {
				span := ss.Spans().At(k)

				// Check if the span has the attribute we want (e.g., "gen_ai.input.messages")
				for _, attrName := range p.config.Attributes {
					if val, ok := span.Attributes().Get(attrName); ok {
						// Assuming the attribute is stored as a string
						textsToAnonymize = append(textsToAnonymize, val.Str())

						// Create a closure so we know exactly how to overwrite this specific attribute later
						pointers = append(pointers, pointer{
							updateFunc: func(cleanText string) {
								val.SetStr(cleanText)
							},
						})
					}
				}
			}
		}
	}

	// Step B: If we found nothing to anonymize, move on to the next processor
	if len(textsToAnonymize) == 0 {
		return traces, nil
	}

	req := client.BatchRequest{
		Texts:    textsToAnonymize,
		Language: "en",
	}

	resp, err := p.client.AnonymizeBatch(ctx, &req)
	if err != nil {
		return traces, err // Handle API failure
	}

	// Step D: Overwrite the original span attributes with the anonymized texts
	for idx, cleanText := range resp.AnonymizedTexts {
		pointers[idx].updateFunc(cleanText)
	}

	return traces, nil
}

type presidioLogsProcessor struct {
	logger            *zap.Logger
	telemetrySettings component.TelemetrySettings
	config            *Config
	nextConsumer      consumer.Logs

	client client.AnonymizerServiceClient
}

func newLogsProcessor(logger *zap.Logger, telemetrySettings component.TelemetrySettings, cfg *Config, nextConsumer consumer.Logs) (*presidioLogsProcessor, error) {
	return &presidioLogsProcessor{
		logger:            logger,
		telemetrySettings: telemetrySettings,
		config:            cfg,
		nextConsumer:      nextConsumer,
	}, nil
}
func (p *presidioLogsProcessor) Start(ctx context.Context, host component.Host) error {
	p.telemetrySettings.Logger.Info("Starting presidio logs processor and building HTTP client")

	// 1. Extract extensions from the host safely
	var extensions map[component.ID]component.Component
	if host != nil {
		extensions = host.GetExtensions()
	}

	// 2. Use the signature you found to build the client
	clientConn, err := p.config.ClientConfig.ToClientConn(
		ctx,
		extensions,
		p.telemetrySettings,
	)
	if err != nil {
		return err // If the client fails to build, the collector will refuse to start
	}
	p.client = client.NewAnonymizerServiceClient(clientConn)

	return nil
}

func (p *presidioLogsProcessor) Shutdown(_ context.Context) error {
	p.logger.Info("Shutting down presidio logs processor")

	return nil
}

func (p *presidioLogsProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true} // Set to true since you'll likely modify logs
}

func (p *presidioLogsProcessor) ConsumeLogs(ctx context.Context, ld plog.Logs) error {

	logs, err := p.processLogs(ctx, ld)
	if err != nil {
		return err
	}

	// 6. Pass the mutated batch to the next processor
	return p.nextConsumer.ConsumeLogs(ctx, logs)
}

func (p *presidioLogsProcessor) processLogs(ctx context.Context, logs plog.Logs) (plog.Logs, error) {
	// Similar structure to processTraces, but adapted for logs instead of traces

	var textsToAnonymize []string

	// We need to keep track of which attribute corresponds to which index in our batch array
	type pointer struct {
		updateFunc func(string)
	}
	var pointers []pointer

	includeLogBody := p.config.IncludeLogBody
	// 1. Iterate through log records and collect texts to anonymize based on config.Attributes

	for i := 0; i < logs.ResourceLogs().Len(); i++ {
		rl := logs.ResourceLogs().At(i)
		for j := 0; j < rl.ScopeLogs().Len(); j++ {
			sl := rl.ScopeLogs().At(j)
			for k := 0; k < sl.LogRecords().Len(); k++ {
				logRecord := sl.LogRecords().At(k)

				if includeLogBody {
					logBody := logRecord.Body().Str()
					textsToAnonymize = append(textsToAnonymize, logBody)
					pointers = append(pointers, pointer{
						updateFunc: func(cleanText string) {
							logRecord.Body().SetStr(cleanText)
						},
					})
				}

				for _, attrName := range p.config.Attributes {
					if val, ok := logRecord.Attributes().Get(attrName); ok {
						textsToAnonymize = append(textsToAnonymize, val.Str())
						pointers = append(pointers, pointer{
							updateFunc: func(cleanText string) {
								val.SetStr(cleanText)
							},
						})
					}
				}
			}
		}
	}
	// 2. Call the anonymization API with the collected texts

	if len(textsToAnonymize) == 0 {
		return logs, nil
	}

	req := client.BatchRequest{
		Texts:    textsToAnonymize,
		Language: "en",
	}

	resp, err := p.client.AnonymizeBatch(ctx, &req)
	if err != nil {
		return logs, err // Handle API failure
	}

	// 3. Update the log records with the anonymized texts
	for idx, cleanText := range resp.AnonymizedTexts {
		pointers[idx].updateFunc(cleanText)
	}
	// 3. Update the log records with the anonymized texts

	return logs, nil
}
