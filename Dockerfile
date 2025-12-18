# Multi-stage Dockerfile for Go application

# Build stage
FROM golang:1.25-alpine AS builder

# Set working directory in the container
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY main.go ./
COPY json/ ./json/
COPY index.html ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests if needed
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy static files if needed
COPY --from=builder /app/json/ ./json/
COPY --from=builder /app/index.html ./index.html

# Expose port (assuming your app listens on port 8080)
EXPOSE 8080

# Command to run the application
CMD ["./main"]