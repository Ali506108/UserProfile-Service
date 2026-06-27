# Stage 1: Build
FROM golang:1.25.11-alpine AS builder

RUN apk add --no-cache git build-base

WORKDIR /app

# Кешируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Сборка бинарника с оптимизациями
RUN go build -ldflags="-s -w" -o server ./cmd/app

# Stage 2: Run
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 9434

CMD ["./server"]
