FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

COPY . .

RUN oapi-codegen -generate types -package api -o internal/api/types.gen.go api/openapi.yaml && \
    oapi-codegen -generate chi-server -package api -o internal/api/server.gen.go api/openapi.yaml

RUN GOOS=linux GOARCH=amd64 go build -o /bin/app ./cmd/app

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /bin/app /bin/app
EXPOSE 8080

ENTRYPOINT ["/bin/app"]
