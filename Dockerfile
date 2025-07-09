# Stage 1 - Build
FROM golang:latest AS builder

WORKDIR /app

# Copy go.mod and go.sum first
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy everything
COPY . ./

# Build the binary, specifying the path to main.go
RUN go build -o matchmaking ./cmd

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/matchmaking .

EXPOSE 8080

ENTRYPOINT ["/app/matchmaking"]
