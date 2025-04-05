FROM golang:1.23 AS builder

# Set working directory inside the container
WORKDIR /app

# Copy application source code
COPY . .
RUN go mod tidy

# Build the Go application
RUN CGO_ENABLED=0 go build -o server ./cmd/auth/main.go

# Use a lightweight Alpine image for production
FROM alpine:latest

WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/server .

# Run the application
CMD ["./server"]
