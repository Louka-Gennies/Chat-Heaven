FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

FROM alpine:latest

RUN apk add --no-cache sqlite

RUN adduser -D myuser

WORKDIR /home/myuser

COPY --from=builder /app/main .
COPY --from=builder /app/static ./static
COPY --from=builder /app/chatHeaven.db .
COPY --from=builder /app/templates ./templates

RUN chown -R myuser:myuser .

USER myuser

EXPOSE 8080

CMD ["./main"]
