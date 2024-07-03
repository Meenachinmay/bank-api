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
COPY sql/schema /app/sql/schema

RUN chmod +x /app/main

# Set environment variables
ARG DB_SOURCE_PROD
ARG DB_SOURCE_TEST
ARG SENDGRID_API

ENV DB_SOURCE_PROD=$DB_SOURCE_PROD
ENV DB_SOURCE_TEST=$DB_SOURCE_TEST
ENV SENDGRID_API=$SENDGRID_API

EXPOSE 8080

CMD ["sh", "-c", "goose -dir /app/sql/schema postgres \"$DB_SOURCE_PROD\" up && /app/main"]