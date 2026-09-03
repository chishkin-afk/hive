FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /bin/app ./cmd/server/main.go

FROM alpine:latest

COPY --from=builder /bin/app app

COPY --from=builder /app/configs/dev.yml /configs/dev.yml
COPY --from=builder /app/internal/web /internal/web
COPY --from=builder /app/.env .env

EXPOSE 8091

CMD ["./app"]