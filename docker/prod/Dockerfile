FROM golang:1.26.1-alpine3.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Compile the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o local_chatbot .

# Use a minimal base image for the final container
FROM alpine:latest AS runner

RUN apk --no-cache add ca-certificates curl wget

WORKDIR /app

COPY --from=builder /app/local_chatbot .

# Copy the static frontend files
COPY --from=builder /app/static ./static

RUN mkdir -p /app/data

# Expose the port the application runs on
EXPOSE 55572

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD [ "curl", "-f", "http://localhost:55572/" ] || exit 1

# Finally run the application
CMD ["./local_chatbot"]
