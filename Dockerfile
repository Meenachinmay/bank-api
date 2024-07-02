# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY . .

# Install goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

RUN go build -o main ./cmd/main.go

# Run stage
FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY .env .
COPY sql/schema /app/sql/schema

RUN chmod +x /app/main

EXPOSE 8080

CMD ["sh", "-c", "goose -dir /app/sql/schema postgres \"$DB_SOURCE\" up && /app/main"]