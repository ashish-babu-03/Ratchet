# Build stage
FROM golang:1.26.5-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY main.go ./
RUN go build -o ratchet main.go

# Run stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/ratchet .
EXPOSE 8080
CMD ["./ratchet"]