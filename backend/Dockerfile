# Multi-stage Dockerfile for Go backend

# Build stage
FROM golang:1.26.5-alpine@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS builder

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

# Final stage
FROM alpine:3.22@sha256:14358309a308569c32bdc37e2e0e9694be33a9d99e68afb0f5ff33cc1f695dce

ARG OCI_REVISION=unknown
ARG OCI_SOURCE=https://git.subcult.tv/subculture-collective/clpr
LABEL org.opencontainers.image.revision="${OCI_REVISION}" \
      org.opencontainers.image.source="${OCI_SOURCE}"

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

RUN addgroup -S -g 10001 clpr && adduser -S -D -H -u 10001 -G clpr clpr \
    && mkdir -p /app/exports && chown -R clpr:clpr /app

WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=clpr:clpr /app/bin/api .

# Copy pSEO HTML templates
COPY --from=builder --chown=clpr:clpr /app/templates ./templates/

USER 10001:10001

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
CMD ["./api"]
