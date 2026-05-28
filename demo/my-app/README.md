# Demo App

This demo app show cases how an OpenAI instrumented application attaches `gen_ai.input.messages` and `gen_ai.output.messages` attributes to the telemetry data, and how the Presidio Processor can be used to detect and redact sensitive information from these attributes.

## Running the demo

### Collector and Infrastructure setup

in the `demo` directory (parent from this one), run the following command to start the OpenTelemetry Collector and the Presidio Processor:

```bash
docker compose up --build
```

### Initial setup

Deps and instrumentation setup [see here](https://opentelemetry.io/docs/zero-code/python/troubleshooting/#bootstrap-using-uv)

```bash
uv sync
uv run opentelemetry-bootstrap -a requirements | uv add --requirement -
```

Copy the `.env.example` file to `.env` and fill in the required OPENAI_API_KEY

```bash
uv run --env-file .env -- opentelemetry-instrument python main.py
```