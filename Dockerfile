# Use the official Golang image as a base
FROM golang:1.22-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN go build -o server .

# Run the server with TLS
CMD ["./server"]