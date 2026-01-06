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

# Copy main application
COPY --from=builder /app/main .
COPY templates ./templates

# Copy seed binaries
RUN mkdir -p /app/scripts/bin
COPY --from=builder /app/seed-gutenberg-books /app/scripts/bin/
COPY --from=builder /app/seed-images /app/scripts/bin/

EXPOSE 8080
CMD ["/app/main"]
