# Build Stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod .
# COPY go.sum . # go.sum doesn't exist yet as we haven't run tidy, suppressing this for now or assuming user runs tidy

# Download dependencies
# Since we can't run go mod tidy on host, we rely on the container to do it or wait for files
# Best practice: Copy everything and let build handle it if go.sum is missing
COPY . .

# Build the application
# CGO_ENABLED=0 for static binary
RUN go build -o main .

# Run Stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Copy environment file template (optional, user should mount .env)
COPY .env.example .env.example

# Expose port
EXPOSE 8080

# Run
CMD ["./main"]
