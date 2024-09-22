FROM golang:1.23.1-alpine3.19 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN apk update && apk upgrade
RUN apk add --no-cache gcc
RUN apk add --no-cache musl-dev
RUN CGO_ENABLED=1 go build -o /app/main .

FROM alpine:3.19
ENV 'DATABASE_URL' 'sqlite3:///app/colablist.db'
ENV 'LISTEN' ':8080'
WORKDIR /app
RUN apk update && apk upgrade
RUN apk add --no-cache sqlite
COPY --from=builder /app/main /app/main
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/static /app/static
RUN sqlite3 -init /app/migrations/0001-schema.sql /app/colablist.db .quit
CMD ["sh" ,"-c" , "/app/main -listen $LISTEN"]



