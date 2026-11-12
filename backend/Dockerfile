# Multi-stage Dockerfile for Go backend

# Build stage
FROM golang:1.26.11-alpine@sha256:ae52b5bc6ed045a9ed26ab03dc19327a802a9cc9aa8b27ed57f5c6b60579b0bb AS builder

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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /app/bin/backfill-canonical-tags ./cmd/backfill-canonical-tags

# Final stage includes the local faster-whisper runtime and ffmpeg. Video is
# streamed from Twitch; only a short-lived normalized WAV is written.
FROM python:3.12-slim-bookworm@sha256:935c3c5b12357a7e10caa9c1e6ca5692344d97cf229832ef584be2cec9fc03df

ARG OCI_REVISION=unknown
ARG OCI_SOURCE=https://git.subcult.tv/subculture-collective/clpr
LABEL org.opencontainers.image.revision="${OCI_REVISION}" \
      org.opencontainers.image.source="${OCI_SOURCE}"

# Install media/runtime dependencies and pin the transcription engine.
# Refresh inherited APT, PCRE2, Perl and OpenSSL packages with Debian security fixes.
RUN apt-get update \
    && apt-get install -y --no-install-recommends apt libapt-pkg6.0 ca-certificates ffmpeg wget libpcre2-8-0 perl-base libssl3 openssl \
    && dpkg --compare-versions "$(dpkg-query -W -f='${Version}' apt)" ge '2.6.2' \
    && dpkg --compare-versions "$(dpkg-query -W -f='${Version}' libapt-pkg6.0)" ge '2.6.2' \
    && dpkg --compare-versions "$(dpkg-query -W -f='${Version}' libssl3)" ge '3.0.22-1~deb12u2' \
    && dpkg --compare-versions "$(dpkg-query -W -f='${Version}' openssl)" ge '3.0.22-1~deb12u2' \
    && pip install --no-cache-dir faster-whisper==1.2.0 av==18.1.0 requests==2.34.2 \
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
COPY --from=builder --chown=clpr:clpr /app/bin/backfill-canonical-tags .

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
