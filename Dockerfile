# Build stage
FROM golang:1.24-alpine AS builder
ENV CGO_ENABLED=1
RUN apk update && apk add --no-cache sqlite-dev gcc musl-dev
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
COPY migrations migrations
RUN go build -o main ./cmd/app

# Run stage
FROM alpine:3.16
RUN apk add --no-cache sqlite
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./main"] 