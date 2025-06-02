# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary using Makefile
RUN make build

# Final stage
FROM scratch

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy the binary
COPY --from=builder /build/build/hartea /usr/local/bin/hartea

# Create a non-root user
USER nobody

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/hartea"]

# Metadata
LABEL org.opencontainers.image.title="HAR Analyzer" \
      org.opencontainers.image.description="Advanced terminal-based HAR file analysis tool with interactive TUI, performance metrics, and professional reporting" \
      org.opencontainers.image.vendor="HAR Analyzer" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.source="https://github.com/GITHUB_REPOSITORY_OWNER/har-analyzer" \
      org.opencontainers.image.documentation="https://github.com/GITHUB_REPOSITORY_OWNER/har-analyzer#readme"