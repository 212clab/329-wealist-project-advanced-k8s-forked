# =============================================================================
# Ops Service Dockerfile (Multi-stage build)
# =============================================================================
# Build from project root: docker build -f services/ops-service/docker/Dockerfile .

# -----------------------------------------------------------------------------
# Stage 1: Build
# -----------------------------------------------------------------------------
FROM golang:1.24-bookworm AS builder

WORKDIR /workspace

# Copy shared package first (for better cache)
COPY packages/wealist-advanced-go-pkg/ ./packages/wealist-advanced-go-pkg/

# Copy ops-service
WORKDIR /workspace/services/ops-service
COPY services/ops-service/ ./

# Add replace directive for shared package
RUN echo 'replace github.com/OrangesCloud/wealist-advanced-go-pkg => ../../packages/wealist-advanced-go-pkg' >> go.mod

# Download dependencies and build
RUN GOWORK=off go mod tidy && GOWORK=off go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOWORK=off go build -o app ./cmd/api

# -----------------------------------------------------------------------------
# Stage 2: Runtime
# -----------------------------------------------------------------------------
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy binary
COPY --from=builder /workspace/services/ops-service/app .

# Copy config (optional, can be overridden by env vars)
COPY --from=builder /workspace/services/ops-service/configs ./configs

# Expose port
EXPOSE 8005

# Run as non-root user
USER nonroot:nonroot

# Entry point
ENTRYPOINT ["/app/app"]
