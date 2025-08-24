# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build optimized static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o miko-manifest .

# Final stage - minimal image
FROM alpine:3.19

# Install only essential certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user for security
RUN adduser -D -H -h /app appuser

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/miko-manifest .

# Switch to non-root user
USER appuser

# Default entrypoint
ENTRYPOINT ["/app/miko-manifest"]
CMD ["--help"]
