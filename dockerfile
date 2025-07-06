# Build Stage
FROM golang:1.23.1-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git protoc protobuf-dev make

# Install protoc plugins for Go
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Set PATH to include Go binaries
ENV PATH="$PATH:/root/go/bin"

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download && go get -u google.golang.org/grpc && go get -u google.golang.org/protobuf

# Copy source code
COPY . .

# Generate protobuf files
RUN make gen-proto

# Generate swagger documentation
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN make gen-docs

# Build the application
RUN go build -o thinkink-server ./cmd/main.go

# Runtime Stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/thinkink-server /app/thinkink-server

# Copy generated files
COPY --from=builder /app/docs /app/docs
COPY --from=builder /app/proto-gen /app/proto-gen

# Create directory for uploaded files
RUN mkdir -p /app/uploads

# Default environment variables - these can be overridden at runtime
ENV PORT=8080 \
    GRPC_PORT=50051 \
    APP_ENV=production \
    DB_HOST=postgres \
    DB_USER=postgres \
    DB_PASSWORD=postgres \
    DB_NAME=postgres \
    DB_PORT=5432 \
    DB_SSL_MODE=disable

# Expose ports
EXPOSE 8080 50051

# Command to run
ENTRYPOINT ["/app/thinkink-server"]