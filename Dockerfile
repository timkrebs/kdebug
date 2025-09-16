# Multi-stage build for kdebug
FROM golang:1.24-alpine AS builder

# Install git and ca-certificates (needed for Go modules)
RUN apk add --no-cache git ca-certificates tzdata

# Create a non-root user
RUN adduser -D -g '' kdebug

# Set working directory
WORKDIR /src

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build arguments
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-w -s -X kdebug/cmd.Version=${VERSION} -X kdebug/cmd.Commit=${COMMIT} -X kdebug/cmd.BuildDate=${DATE}" \
    -a -installsuffix cgo \
    -o kdebug \
    .

# Final stage: minimal image
FROM scratch

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy user
COPY --from=builder /etc/passwd /etc/passwd

# Copy binary
COPY --from=builder /src/kdebug /usr/local/bin/kdebug

# Use non-root user
USER kdebug

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/kdebug"]

# Default command
CMD ["--help"]

# Labels
LABEL org.opencontainers.image.title="kdebug"
LABEL org.opencontainers.image.description="A CLI tool that automatically diagnoses common Kubernetes issues"
LABEL org.opencontainers.image.vendor="kdebug"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/timkrebs/kdebug"
