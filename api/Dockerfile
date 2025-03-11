# Builder stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install necessary build tools
RUN apk add --no-cache git

# Copy go.mod and go.sum files first to leverage Docker cache
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o shopware-shopify-api ./cmd/api

# Final stage
FROM alpine:latest

# Add necessary runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user to run the application
RUN adduser -D -H -h /app appuser

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/shopware-shopify-api .

# Copy the environment file
COPY .env .

# Set the ownership of the application files
RUN chown -R appuser:appuser /app

# Use the non-root user
USER appuser

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./shopware-shopify-api"]