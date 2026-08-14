# Multi-stage Dockerfile for Go backend

# Build stage
FROM golang:1.26.6-alpine@sha256:af8d6740070b8906d12eae1c3e3ea0957fb63f492051ea05e354c38ef9fe88df AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /app/bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /app/bin/seed-site-freshness ./cmd/seed-site-freshness
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /app/bin/backfill-clip-topics ./cmd/backfill-clip-topics

# Final stage includes the local faster-whisper runtime and ffmpeg. Video is
# streamed from Twitch; only a short-lived normalized WAV is written.
FROM python:3.12-slim-bookworm@sha256:4766d8b510c428e595d74b9cc5bbb2fae8e26316fffb4adc89908d79aacd58a2

ARG OCI_REVISION=unknown
ARG OCI_SOURCE=https://git.subcult.tv/subculture-collective/clpr
LABEL org.opencontainers.image.revision="${OCI_REVISION}" \
      org.opencontainers.image.source="${OCI_SOURCE}"

# Install media/runtime dependencies and pin the transcription engine.
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates ffmpeg wget \
    && pip install --no-cache-dir faster-whisper==1.2.0 requests==2.34.2 \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd --system --gid 10001 clpr \
    && useradd --system --uid 10001 --gid 10001 --no-create-home clpr \
    && mkdir -p /app/exports /app/whisper-models \
    && chown -R clpr:clpr /app

WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=clpr:clpr /app/bin/api .
COPY --from=builder --chown=clpr:clpr /app/bin/seed-site-freshness .
COPY --from=builder --chown=clpr:clpr /app/bin/backfill-clip-topics .

# Copy pSEO HTML templates
COPY --from=builder --chown=clpr:clpr /app/templates ./templates/

# Copy the self-contained faster-whisper runner.
COPY --from=builder --chown=clpr:clpr /app/whisper ./whisper/

ENV WHISPER_PYTHON_PATH=/usr/local/bin/python3 \
    WHISPER_RUNNER_DIR=/app/whisper \
    WHISPER_FFMPEG_PATH=/usr/bin/ffmpeg \
    WHISPER_MODEL_DIR=/app/whisper-models \
    HF_HOME=/app/whisper-models/huggingface \
    XDG_CACHE_HOME=/app/whisper-models/cache

USER 10001:10001

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO /dev/null http://localhost:8080/api/v1/health || exit 1

# Run the binary
CMD ["./api"]
