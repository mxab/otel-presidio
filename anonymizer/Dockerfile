# Install uv
FROM python:3.12-slim AS builder
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

# Use the system Python across both stages
ENV UV_PYTHON_DOWNLOADS=0
ENV UV_COMPILE_BYTECODE=1
# Change the working directory to the `app` directory
WORKDIR /app

# Install dependencies
RUN --mount=type=cache,target=/root/.cache/uv \
    --mount=type=bind,source=uv.lock,target=uv.lock \
    --mount=type=bind,source=pyproject.toml,target=pyproject.toml \
    uv sync --locked --no-install-project --no-editable

RUN --mount=type=cache,target=/root/.cache/uv \
    --mount=type=bind,source=uv.lock,target=uv.lock \
    --mount=type=bind,source=pyproject.toml,target=pyproject.toml \
    uv run spacy download en_core_web_lg

# Copy the project into the intermediate image
COPY . /app

# Sync the project
# RUN --mount=type=cache,target=/root/.cache/uv \
#     uv sync --locked --no-editable


FROM python:3.12-slim

# Copy the environment, but not the source code
COPY --from=builder /app /app


# Run the application
CMD ["/app/.venv/bin/python", "/app/main.py"]