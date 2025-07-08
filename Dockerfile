# Start from the official Golang image
FROM golang:latest

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum first (to leverage Docker cache)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of your code
COPY . .

# Build the Go binary
RUN go build -o matchmaking ./cmd

# Expose the port (make sure it matches SERVER_PORT)
EXPOSE 8080

# Command to run the binary
CMD ["./matchmaking"]
