# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
ENV GOSUMDB=off
RUN go mod download
COPY . .

# Build main application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/web

# Build seed binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/seed-gutenberg-books ./scripts/seed-gutenberg-books.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/seed-images ./scripts/seed-images.go

# Final stage
FROM alpine:latest
WORKDIR /app

# Install CA certificates for HTTPS requests (needed for downloading Gutenberg covers)
# Use HTTP for apk repos to bootstrap ca-certificates, then restore HTTPS
RUN sed -i 's/https/http/' /etc/apk/repositories && \
    apk update && \
    apk add --no-cache ca-certificates && \
    update-ca-certificates && \
    sed -i 's/http/https/' /etc/apk/repositories

# Set SSL cert path for Go binaries (statically compiled binaries need this)
ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

# Copy main application
COPY --from=builder /app/main .
COPY templates ./templates

# Copy seed binaries
RUN mkdir -p /app/scripts/bin
COPY --from=builder /app/seed-gutenberg-books /app/scripts/bin/
COPY --from=builder /app/seed-images /app/scripts/bin/

# Copy migration files (for init jobs)
RUN mkdir -p /app/migrations
COPY migrations/*.sql /app/migrations/

EXPOSE 8080
CMD ["/app/main"]
