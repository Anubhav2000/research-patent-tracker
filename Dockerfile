FROM golang:1.21-alpine

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/api

# Create a new stage with minimal image
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=0 /app/main .
COPY --from=0 /app/.env .

EXPOSE 8080

CMD ["./main"] 