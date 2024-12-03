FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o migrator cmd/migrator/main.go
RUN go build -o songs-lib cmd/songs-lib/main.go

FROM alpine:latest

RUN apk update && apk --no-cache add \
    postgresql-client

WORKDIR /root/

COPY --from=builder /app/migrator .
COPY --from=builder /app/songs-lib .
COPY ./migrations /root/migrations
COPY .env .env
COPY wait-for-postgres.sh /root/

RUN chmod +x /root/wait-for-postgres.sh

CMD ["./songs-lib"]