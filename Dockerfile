# Start from the official Golang image
FROM golang:1.24-alpine AS builder

# Install required packages
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o album-store-plus ./cmd/api

# Create a minimal production image
FROM alpine:latest

# Install necessary runtime packages
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/album-store-plus .

# Expose the application port
EXPOSE 8080

# Run the binary
CMD ["./album-store-plus"]