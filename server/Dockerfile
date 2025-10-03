# ============================================
# Development Stage
# ============================================
FROM golang:1.24-alpine AS development

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


RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# ============================================
# Production Stage
# ============================================
FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/


COPY --from=builder /app/main .


COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./main"]