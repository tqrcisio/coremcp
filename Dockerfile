# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o coremcp ./cmd/coremcp

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS connections
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/coremcp .

# Copy example config and use it as the default config
COPY coremcp.example.yaml /app/coremcp.example.yaml
COPY coremcp.example.yaml /app/coremcp.yaml

# Create non-root user
RUN addgroup -g 1000 coremcp && \
    adduser -D -u 1000 -G coremcp coremcp && \
    chown -R coremcp:coremcp /app

USER coremcp

# Default command
ENTRYPOINT ["./coremcp"]
CMD ["serve"]
