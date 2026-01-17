# Dockerfile for nocturnal + OpenCode
# Multi-stage build for CI/CD agent execution

# Build nocturnal from source
FROM golang:1.23-alpine AS nocturnal-builder

RUN apk add --no-cache git

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG BUILD_TIME
RUN BUILD_TIME=${BUILD_TIME:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")} && \
    go build -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -w -s" \
    -o nocturnal .

# Final stage with Node.js for OpenCode
FROM node:22-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    git \
    curl \
    jq \
    && rm -rf /var/lib/apt/lists/*

# Install OpenCode globally
RUN npm install -g opencode-ai

# Copy nocturnal binary
COPY --from=nocturnal-builder /build/nocturnal /usr/local/bin/nocturnal
RUN chmod +x /usr/local/bin/nocturnal

# Configure git for CI use
RUN git config --global user.email "opencode@ci.local" && \
    git config --global user.name "OpenCode CI"

# Create workspace directory
WORKDIR /workspace

# Verify installations
RUN nocturnal --version && opencode --version

# Default: show help
CMD ["sh", "-c", "echo 'nocturnal + opencode ready' && nocturnal --version && opencode --version"]
