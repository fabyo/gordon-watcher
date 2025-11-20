# ═══════════════════════════════════════════════════════════
#  GORDON WATCHER - PRODUCTION DOCKERFILE (Multi-stage)
# ═══════════════════════════════════════════════════════════

# ─────────────────────────────────────────────────────────
#  STAGE 1: Builder
# ─────────────────────────────────────────────────────────
FROM golang:1.21-alpine AS builder

# Build arguments
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies (cached if go.mod/go.sum unchanged)
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s \
    -X main.version=${VERSION} \
    -X main.commit=${COMMIT} \
    -X main.buildDate=${BUILD_DATE}" \
    -o gordon-watcher \
    ./cmd/watcher

# ─────────────────────────────────────────────────────────
#  STAGE 2: Runtime (minimal)
# ─────────────────────────────────────────────────────────
FROM alpine:latest

# Metadata
LABEL maintainer="fabyo <fabyo@example.com>"
LABEL org.opencontainers.image.source="https://github.com/fabyo/gordon-watcher"
LABEL org.opencontainers.image.description="Gordon Watcher - File System Monitor"
LABEL org.opencontainers.image.licenses="MIT"

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1000 gordon && \
    adduser -D -u 1000 -G gordon gordon

# Create directories
RUN mkdir -p /opt/gordon-watcher/bin \
             /opt/gordon-watcher/data \
             /opt/gordon-watcher/logs \
             /etc/gordon/watcher && \
    chown -R gordon:gordon /opt/gordon-watcher /etc/gordon

# Copy binary from builder
COPY --from=builder --chown=gordon:gordon /build/gordon-watcher /opt/gordon-watcher/bin/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Set timezone (Brazil)
ENV TZ=America/Sao_Paulo

# Switch to non-root user
USER gordon

# Set working directory
WORKDIR /opt/gordon-watcher

# Expose ports
EXPOSE 8081 9100

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8081/health || exit 1

# Entry point
ENTRYPOINT ["/opt/gordon-watcher/bin/gordon-watcher"]
