# Start from the official Golang image
FROM golang:latest as builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum first (to leverage Docker cache)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of your code
COPY . .

# Stage 2 - Minimal runtime
FROM gcr.io/distroless/base-debian11

WORKDIR /app

COPY --from=builder /app/matchmaking .

# Expose the port your app uses
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/app/matchmaking"]