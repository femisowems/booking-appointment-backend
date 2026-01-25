# Build Stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 disables CGO for a static binary
# -o api sets the output binary name
RUN CGO_ENABLED=0 GOOS=linux go build -o api cmd/api/main.go

# Run Stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies (e.g. ca-certificates for HTTPS)
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /app/api .
# Copy migrations if needed (or rely on them being embedded/external, but code looks for file system)
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./api"]
