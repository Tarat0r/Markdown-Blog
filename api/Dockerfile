# Dockerfile для /api
FROM golang:1.24 AS builder

WORKDIR /app

COPY api/go.mod api/go.sum ./
RUN go mod download

COPY api/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /markdown-blog main.go


# Установка зависимостей
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /markdown-blog .
COPY database/markdown_blog.sql /database/markdown_blog.sql

# Установка переменной окружения
ENV HOST_ADDRESS=0.0.0.0:8080

ENV DATABASE_URL=postgres://admin:admin@host.docker.internal:5432/postgres?sslmode=disable
ENV STATIC_PATH=images

# Установка переменной окружения для JWT
EXPOSE 8080
CMD ["./markdown-blog"]
