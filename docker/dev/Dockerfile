FROM golang:1.26.1-alpine3.23

# Install air for hot reloading
RUN go install github.com/air-verse/air@latest

# Install additional tools
RUN apk add --no-cache curl wget ca-certificates

WORKDIR /app

# Copy source code
COPY . .

# Download dependencies
RUN go mod download

# Expose the port
EXPOSE 55572

# Run with air for hot reload
CMD ["air", "-c", ".air.toml"]
