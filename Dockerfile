FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

FROM builder AS tester
RUN go test -v ./...

FROM builder AS compiler
RUN go build -o main ./cmd/server/server.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=compiler /app/main .
COPY --from=compiler /app/migrations ./migrations

EXPOSE 8080

CMD ["./main"]