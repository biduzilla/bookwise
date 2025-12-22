FROM golang:1.25.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/bin/api ./cmd/api

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/bin/migrate ./cmd/migrate

FROM alpine:3.18

WORKDIR /root/

COPY --from=builder /app/bin/api ./bookwise
COPY --from=builder /app/bin/migrate ./migrate

COPY --from=builder /app/migrations/*.sql ./migrations/

RUN chmod +x /root/bookwise /root/migrate \
    && apk --no-cache add tzdata

ENV SERVER_PORT=4000 \
    MIGRATE_PATH=/root/migrations

ENTRYPOINT ["/root/bookwise"]