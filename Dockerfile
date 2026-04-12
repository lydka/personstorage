FROM golang:1.22-alpine AS builder

RUN apk add --no-cache build-base

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /out/personstorage .

FROM alpine:3.20

RUN adduser -D -h /app appuser \
	&& mkdir -p /app/data \
	&& chown -R appuser:appuser /app

WORKDIR /app

COPY --from=builder /out/personstorage /app/personstorage

ENV LISTEN_ADDR=:8080
ENV DATABASE_PATH=/app/data/app.db

EXPOSE 8080

USER appuser

ENTRYPOINT ["/app/personstorage"]
