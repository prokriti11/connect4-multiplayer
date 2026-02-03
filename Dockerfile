# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy frontend files
COPY --from=builder /app/frontend ./frontend

# Expose port
EXPOSE 8080

# Set environment variables
ENV PORT=8080

# Run the server
CMD ["./server"]
