# Этап 1: Сборка приложения
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./spotService/cmd/main.go

# Этап 2: Финальный образ (минимальный)
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем собранное приложение
COPY --from=builder /app/main .

# Запускаем
CMD ["./main"]