# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
ENV GOSUMDB=off
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/web

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY templates ./templates
EXPOSE 8080
CMD ["/app/main"]
