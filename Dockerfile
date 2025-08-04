# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies for building
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Install templ for template generation
RUN go install github.com/a-h/templ/cmd/templ@latest

# Generate templates
RUN templ generate

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o socgo .

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Create directories for volumes
RUN mkdir -p /app/data /app/uploads

# Copy the binary from builder stage
COPY --from=builder /app/socgo .

# Copy web assets
COPY --from=builder /app/web ./web

# Create non-root user for security
RUN addgroup -g 1000 socgo && \
    adduser -D -s /bin/sh -u 1000 -G socgo socgo && \
    chown -R socgo:socgo /app

USER socgo

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8081/ || exit 1

# Run the application
CMD ["./socgo"]