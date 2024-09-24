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
WORKDIR /app
RUN mkdir /app/data
RUN apk update && apk upgrade
RUN apk add --no-cache sqlite
COPY --from=builder /app/main /app/main
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/static /app/static
CMD ["sh" ,"-c" , "/app/main", "-listen", ":8080", "-database-url", "/app/data/database.db"]
