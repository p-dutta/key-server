# Start from a Golang base image for building
FROM golang:latest as builder

# Set the working directory inside the container
WORKDIR /go/src/ksm

# Install hot reload tool
# RUN go install github.com/cosmtrek/air@latest

COPY .env .

# Copy go.mod and go.sum files to the working directory
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source code to the working directory
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -a -installsuffix cgo -o key-server .

# Use a minimal base image for the final build
FROM alpine:latest AS final

# Set the working directory inside the container
WORKDIR /app

# Update the package list and install curl
RUN apk update && apk add --no-cache curl

# Copy the compiled binary from the builder stage to the current directory in the container
COPY --from=builder /go/src/ksm/key-server .

# Copy the .env file and any other required files to the container
COPY .env .
# COPY *.pem .

ENV TZ=Asia/Dhaka

# Expose the port the application listens on
EXPOSE 8080

# Command to run the application
CMD ["./key-server"]
