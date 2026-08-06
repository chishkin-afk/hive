FROM golang:1.25.5 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o hiveback cmd/server/main.go

# ---
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/configs/ ./configs/
COPY --from=builder /app/hiveback .
COPY --from=builder /app/docs/ ./docs/
COPY --from=builder /app/migrations/ ./migrations/

RUN chmod +x /app/hiveback
RUN apk add --no-cache tzdata

EXPOSE 8090

CMD [ "./hiveback" ]