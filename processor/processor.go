package presidioprocessor

import (
	"context"
	"net/http"

	"github.com/mxab/presidio-processor/client"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

type presidioProcessor struct {
	logger            *zap.Logger
	telemetrySettings component.TelemetrySettings
	config            *Config
	nextConsumer      consumer.Traces
	httpClient        *http.Client // This is your custom client for calling Presidio

	client client.AnonymizerServiceClient
}

func newTracesProcessor(logger *zap.Logger, telemetrySettings component.TelemetrySettings, cfg *Config, nextConsumer consumer.Traces) (*presidioProcessor, error) {
	return &presidioProcessor{
		logger:            logger,
		telemetrySettings: telemetrySettings,
		config:            cfg,
		nextConsumer:      nextConsumer,
	}, nil
}

func (p *presidioProcessor) Start(ctx context.Context, host component.Host) error {
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

func (p *presidioProcessor) Shutdown(_ context.Context) error {
	p.logger.Info("Shutting down presidio processor")

	return nil
}

func (p *presidioProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true} // Set to true since you'll likely modify traces
}

func (p *presidioProcessor) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {

	traces, err := p.processTraces(ctx, td)
	if err != nil {
		return err
	}

	// 6. Pass the mutated batch to the next processor
	return p.nextConsumer.ConsumeTraces(ctx, traces)
}

// 2. The core function that OTel calls for every batch of traces
func (p *presidioProcessor) processTraces(ctx context.Context, traces ptrace.Traces) (ptrace.Traces, error) {
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
