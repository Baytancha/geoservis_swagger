# Используем официальный образ Go как базовый
FROM golang:latest as builder


WORKDIR /app
ENV GOOS=linux

ENV CGO_ENABLED=0
COPY . .
RUN go mod tidy

# Собираем приложение
RUN go build -o main


FROM alpine:latest
WORKDIR /app


COPY --from=builder /app/main .
EXPOSE 8080


CMD ["./main"]
