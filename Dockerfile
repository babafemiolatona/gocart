FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o gocart ./cmd/api

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates \
    && adduser -D -u 10001 gocart

COPY --from=builder /app/gocart .

RUN chown -R gocart:gocart /app

USER gocart

EXPOSE 8080

CMD ["./gocart"]