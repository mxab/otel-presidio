# THE CONTEXT OF THIS DOCKERIFLE WILL BE THE ROOT OF THE REPO, NOT THE demo/collector FOLDER
FROM alpine:3.19 AS certs
RUN apk --update add ca-certificates

FROM golang:1.26.2 AS build-stage
WORKDIR /build
# THE CONTEXT OF THIS DOCKERIFLE WILL BE THE ROOT OF THE REPO, NOT THE demo/collector FOLDER
COPY demo/collector/builder-config.yaml .

RUN --mount=type=cache,target=/root/.cache/go-build GO111MODULE=on go install go.opentelemetry.io/collector/cmd/builder@v0.152.0

COPY processor/go.mod processor/go.sum ./
COPY processor ./processor

RUN --mount=type=cache,target=/root/.cache/go-build builder --config builder-config.yaml

FROM gcr.io/distroless/base:latest

ARG USER_UID=10001
USER ${USER_UID}

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --chmod=755 --from=build-stage /build/otelcol-dev /otelcol
VOLUME /otelcol/config
ENTRYPOINT ["/otelcol/otelcol-dev"]
CMD ["--config", "/otelcol/config/collector-config.yaml"]

EXPOSE 4317 4318 12001
