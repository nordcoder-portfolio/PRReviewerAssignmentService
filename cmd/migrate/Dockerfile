FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o /bin/migrate ./cmd/migrate

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /bin/migrate /bin/migrate
COPY migrations /migrations

ENTRYPOINT ["/bin/migrate"]
