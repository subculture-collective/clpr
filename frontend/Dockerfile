# Multi-stage Dockerfile for React frontend

# Build stage
FROM node:24-alpine@sha256:a0b9bf06e4e6193cf7a0f58816cc935ff8c2a908f81e6f1a95432d679c54fbfd AS builder

WORKDIR /app

# Copy package files
COPY frontend/package.json frontend/package-lock.json ./

# Install dependencies
RUN npm ci

# Copy source code and environment file
COPY frontend/ .
COPY docs/openapi/generated /app/public/docs/openapi/generated

# Use production environment for build
ARG NODE_ENV=production
ENV NODE_ENV=${NODE_ENV}
ARG VITE_API_BASE_URL=/api/v1
ARG VITE_BASE_URL=/
ARG VITE_PUBLIC_SITE_URL=https://clpr.tv
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
ENV VITE_BASE_URL=${VITE_BASE_URL}
ENV VITE_PUBLIC_SITE_URL=${VITE_PUBLIC_SITE_URL}

RUN case "$VITE_PUBLIC_SITE_URL" in \
      http://*|https://*) ;; \
      *) echo "VITE_PUBLIC_SITE_URL must be an absolute HTTP(S) origin" >&2; exit 1 ;; \
    esac

# Remove dev .env file so .env.production is used
RUN rm -f .env .env.local

# Build the application (Vite will use .env.production automatically)
RUN npm run routes:check && npm run routes:nginx && npm run build

# Production stage
FROM nginx:1.29-alpine@sha256:5616878291a2eed594aee8db4dade5878cf7edcb475e59193904b198d9b830de

ARG OCI_REVISION=unknown
ARG OCI_SOURCE=https://git.subcult.tv/subculture-collective/clpr
LABEL org.opencontainers.image.revision="${OCI_REVISION}" \
      org.opencontainers.image.source="${OCI_SOURCE}"

# Copy custom nginx config
COPY frontend/nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /app/edge-routes.conf /etc/nginx/edge-routes.conf

# Copy built assets from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Static assets may retain restrictive source modes from the build context.
# nginx runs as an unprivileged user and must be able to traverse and read all
# published files.
RUN chmod -R a+rX /usr/share/nginx/html

# Preserve the favicon asset's canonical source while also serving the
# conventional root URL expected by browsers and crawlers.
RUN cp /usr/share/nginx/html/favicon_io/favicon.ico /usr/share/nginx/html/favicon.ico

# Apply security updates available for the pinned base snapshot.
RUN apk upgrade --no-cache

USER 101:101

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:8080/health.html || exit 1

# Run nginx
CMD ["nginx", "-g", "daemon off;"]
