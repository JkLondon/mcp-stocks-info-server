# Этап сборки
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Установка зависимостей для сборки
RUN apk add --no-cache git

# Копирование и загрузка зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o mcp-stocks-server ./cmd/server

# Финальный образ
FROM alpine:3.18

WORKDIR /app

# Установка необходимых пакетов
RUN apk add --no-cache ca-certificates tzdata

# Копирование исполняемого файла из этапа сборки
COPY --from=builder /app/mcp-stocks-server .
COPY --from=builder /app/config.yaml.example ./config.yaml

# Порт, который будет использоваться приложением
EXPOSE 8080

# Запуск приложения с указанным конфигурационным файлом
CMD ["./mcp-stocks-server", "./config.yaml"] 