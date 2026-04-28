# Stage 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o insighta .

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

# Install root CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/insighta .

# Create a volume for credentials
VOLUME ["/root/.insighta"]

# Set the entrypoint
ENTRYPOINT ["./insighta"]
