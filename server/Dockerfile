# ============================================
# Development Stage
# ============================================
FROM golang:1.25-alpine AS development

# Install air for hot reload
RUN go install github.com/air-verse/air@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

# ============================================
# Builder Stage
# ============================================
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Corriger le chemin de build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

# ============================================
# Production Stage
# ============================================
FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/main .
# Corriger le chemin des migrations
COPY --from=builder /app/internal/migrations ./migrations

EXPOSE 8080

CMD ["./main"]