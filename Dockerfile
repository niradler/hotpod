# Build stage
FROM golang:1.23-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o agent ./agent/

# Final stage
FROM python:3.9-alpine

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/agent .

# Command to run the executable
CMD ["./agent", "-command", "python -m http.server 8000", "-port", "8080", "-keepAlive", "-host", ""]