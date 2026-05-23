#!/bin/bash


echo "starting app with OpenTelemetry instrumentation..."

if [ -f .env ]; then
  # Automatically export all variables defined in the file
  set -a
  source .env
  set +a
else
  echo "Error: .env file not found."
  exit 1
fi

uv run opentelemetry-instrument python main.py
