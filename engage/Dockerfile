FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /bin/app ./cmd/main.go

FROM alpine:latest

COPY --from=builder /bin/app app

COPY --from=builder /app/migrations /migrations
COPY --from=builder /app/config/config.yml /config/config.yml
COPY --from=builder /app/.env .env

EXPOSE 8041

CMD ["./app"]