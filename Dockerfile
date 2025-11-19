FROM golang:1.24 AS builder
WORKDIR /app

# Копируем только go.mod (go.sum может быть пустым)
COPY go.mod ./
RUN go mod tidy

# Копируем весь проект и собираем бинарник
COPY . .
RUN go build -o app .

# --- Рантайм ---
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/app .
EXPOSE 8080
CMD ["./app"]
