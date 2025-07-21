# Build stage
FROM golang:1.24-alpine AS builder

# Install git for Go modules
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o miko-manifest .

# Final stage
FROM alpine:latest

# Install yamllint and other dependencies
RUN apk add --no-cache \
    python3 \
    py3-pip \
    && pip3 install --break-system-packages yamllint

# Create non-root user
RUN addgroup -g 1001 -S miko && \
    adduser -u 1001 -S miko -G miko

# Copy the binary from builder stage
COPY --from=builder /app/miko-manifest /usr/local/bin/miko-manifest

# Set permissions
RUN chmod +x /usr/local/bin/miko-manifest

# Switch to non-root user
USER miko

# Set working directory
WORKDIR /workspace

# Default command
ENTRYPOINT ["miko-manifest"]
CMD ["--help"]
