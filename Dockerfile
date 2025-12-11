# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod .
COPY main.go .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway main.go

# Runtime stage
FROM alpine:3.18

RUN apk add --no-cache ca-certificates
WORKDIR /root/

COPY --from=builder /app/api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]
