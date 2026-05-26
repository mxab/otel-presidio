package presidioprocessor

import (
	"path/filepath"
	"testing"

	"github.com/moby/moby/api/types/build"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestTracesProcessor(t *testing.T) {

	endpoint := anonymizerConnection(t)
	clientConfig := configgrpc.NewDefaultClientConfig()
	clientConfig.Endpoint = endpoint // Ensure this matches the port your Presidio anonymizer is listening on
	clientConfig.TLS.Insecure = true // Disable TLS for testing; ensure your Presidio server is configured accordingly
	config := &Config{
		ClientConfig: clientConfig,
		Attributes:   []string{"foo.bar", "flim.flam"},
	}

	sink := new(consumertest.TracesSink)

	telemetrySettings := component.TelemetrySettings{
		Logger: zap.NewNop(),
	}
	processor, err := newTracesProcessor(zap.NewNop(), telemetrySettings, config, sink)
	require.NoError(t, err)

	err = processor.Start(t.Context(), nil)
	require.NoError(t, err)

	// Create test traces
	td := ptrace.NewTraces()
	rs := td.ResourceSpans().AppendEmpty()
	rs.Resource().Attributes().PutStr("service.name", "checkout-canary")

	span1 := rs.ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	span1.SetName("test-span")
	span1.Attributes().PutStr("foo.bar", "My name is Max and I have a credit card number 4111 1111 1111 1111")
	span1.Attributes().PutStr("something.else", "My name is Max and I have a credit card number 4111 1111 1111 1111")

	span2 := rs.ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	span2.SetName("test-span2")
	span2.Attributes().PutStr("foo.bar", "I live in New York and my email is max@example.com")
	span2.Attributes().PutStr("flim.flam", "This is my address: 123 Main St, New York, NY")

	// Process traces
	ctx := t.Context()
	err = processor.ConsumeTraces(ctx, td)
	require.NoError(t, err)

	// Verify results
	require.Equal(t, 2, sink.SpanCount())

	processedSpan := sink.AllTraces()[0].ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0)

	span1FooBar, ok := processedSpan.Attributes().Get("foo.bar")
	assert.True(t, ok)
	assert.Equal(t, "My name is <PERSON> and I have a credit card number <CREDIT_CARD>", span1FooBar.Str())
	span1SomethingElse, ok := processedSpan.Attributes().Get("something.else")
	assert.True(t, ok)
	assert.Equal(t, "My name is Max and I have a credit card number 4111 1111 1111 1111", span1SomethingElse.Str())

	processedSpan2 := sink.AllTraces()[0].ResourceSpans().At(0).ScopeSpans().At(1).Spans().At(0)
	span2FooBar, ok := processedSpan2.Attributes().Get("foo.bar")
	assert.True(t, ok)
	assert.Equal(t, "I live in <LOCATION> and my email is <EMAIL_ADDRESS>", span2FooBar.Str())
	span2FlimFlam, ok := processedSpan2.Attributes().Get("flim.flam")
	assert.True(t, ok)
	assert.Equal(t, "This is my address: 123 <LOCATION>, <LOCATION>, <LOCATION>", span2FlimFlam.Str())

}

// func BenchmarkDeploymentProcessor(b *testing.B) {
// 	config := &Config{
// 		ServicePatterns: map[string]DeploymentInfo{
// 			"service-[0-9]+": {Environment: "prod"},
// 		},
// 		DefaultDeployment: DeploymentInfo{Environment: "unknown"},
// 	}

// 	sink := new(consumertest.TracesSink)
// 	processor, _ := newDeploymentProcessor(zap.NewNop(), config, sink)

// 	td := createTestTraces(100) // Helper to create 100 spans

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_ = processor.ConsumeTraces(context.Background(), td)
// 	}
// }

func TestLogsProcessor(t *testing.T) {
	// Similar structure to TestTracesProcessor, but create and verify logs instead of traces
	endpoint := anonymizerConnection(t)
	clientConfig := configgrpc.NewDefaultClientConfig()
	clientConfig.Endpoint = endpoint
	clientConfig.TLS.Insecure = true // Disable TLS for testing; ensure your Presidio server is configured accordingly
	config := &Config{
		ClientConfig:   clientConfig,
		Attributes:     []string{"foo.bar", "flim.flam"},
		IncludeLogBody: false, // Set to true if you want to also anonymize the log body
	}

	sink := new(consumertest.LogsSink)

	telemetrySettings := component.TelemetrySettings{
		Logger: zap.NewNop(),
	}
	processor, err := newLogsProcessor(zap.NewNop(), telemetrySettings, config, sink)
	require.NoError(t, err)

	err = processor.Start(t.Context(), nil)
	require.NoError(t, err)

	ld := plog.NewLogs()
	rl := ld.ResourceLogs().AppendEmpty()
	record := rl.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	record.Body().SetStr("Feel free to contact me via bob@example.com or send a letter to 156 Banana St, Springfield, IL")
	record.Attributes().PutStr("foo.bar", "this is important: 234-567-8901")
	record.Attributes().PutStr("flim.flam", "another important thing: january the first of 2020")

	err = processor.ConsumeLogs(t.Context(), ld)
	require.NoError(t, err)

	require.Equal(t, 1, sink.LogRecordCount())
	processedLog := sink.AllLogs()[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
	body := processedLog.Body().Str()
	// include log body is false, so it should remain unchanged
	assert.Equal(t, "Feel free to contact me via bob@example.com or send a letter to 156 Banana St, Springfield, IL", body)

	fooBar, ok := processedLog.Attributes().Get("foo.bar")
	assert.True(t, ok)
	assert.Equal(t, "this is important: <PHONE_NUMBER>", fooBar.Str())

	flimFlam, ok := processedLog.Attributes().Get("flim.flam")
	assert.True(t, ok)
	assert.Equal(t, "another important thing: <DATE_TIME>", flimFlam.Str())

}
func TestLogsIncludeBodyProcessor(t *testing.T) {
	// Similar structure to TestTracesProcessor, but create and verify logs instead of traces
	endpoint := anonymizerConnection(t)
	clientConfig := configgrpc.NewDefaultClientConfig()
	clientConfig.Endpoint = endpoint
	clientConfig.TLS.Insecure = true // Disable TLS for testing; ensure your Presidio server is configured accordingly
	config := &Config{
		ClientConfig:   clientConfig,
		Attributes:     []string{"foo.bar", "flim.flam"},
		IncludeLogBody: true,
	}

	sink := new(consumertest.LogsSink)

	telemetrySettings := component.TelemetrySettings{
		Logger: zap.NewNop(),
	}
	processor, err := newLogsProcessor(zap.NewNop(), telemetrySettings, config, sink)
	require.NoError(t, err)

	err = processor.Start(t.Context(), nil)
	require.NoError(t, err)

	ld := plog.NewLogs()
	rl := ld.ResourceLogs().AppendEmpty()
	record := rl.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	record.Body().SetStr("Feel free to contact me via bob@example.com or send a letter to 156 Banana St, Springfield, IL")
	record.Attributes().PutStr("foo.bar", "this is important: 234-567-8901")
	record.Attributes().PutStr("flim.flam", "another important thing: january the first of 2020")

	err = processor.ConsumeLogs(t.Context(), ld)
	require.NoError(t, err)

	require.Equal(t, 1, sink.LogRecordCount())
	processedLog := sink.AllLogs()[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
	body := processedLog.Body().Str()
	assert.Equal(t, "Feel free to contact me via <EMAIL_ADDRESS> or send a letter to 156 Banana St, <LOCATION>, <LOCATION>", body)

	fooBar, ok := processedLog.Attributes().Get("foo.bar")
	assert.True(t, ok)
	assert.Equal(t, "this is important: <PHONE_NUMBER>", fooBar.Str())

	flimFlam, ok := processedLog.Attributes().Get("flim.flam")
	assert.True(t, ok)
	assert.Equal(t, "another important thing: <DATE_TIME>", flimFlam.Str())

}

func anonymizerConnection(t *testing.T) (endpoint string) {
	opts := []tc.ContainerCustomizer{
		tc.WithExposedPorts("50051/tcp"),
		tc.WithWaitStrategy(wait.ForLog("gRPC Server started successfully.")),
		tc.WithDockerfile(tc.FromDockerfile{
			Context:    filepath.Join("..", "anonymizer"),
			Dockerfile: "Dockerfile",
			BuildOptionsModifier: func(ibo *client.ImageBuildOptions) {

				ibo.Version = build.BuilderBuildKit
			},
		}),
	}

	ctx := t.Context()
	c, err := tc.Run(ctx, "", opts...)
	require.NoError(t, err)

	endpoint, err = c.Endpoint(ctx, "")
	require.NoError(t, err)

	return endpoint

}
