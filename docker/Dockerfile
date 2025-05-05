# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go source code into the container
COPY . .

# Build the Go binary for Alpine
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pgbase examples/base/main.go

# Stage 2: Create the final image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /pb

# Copy the pgbase binary from the builder stage
COPY --from=builder /app/pgbase .

# Copy .env file into the container
COPY docker/.env .

# Ensure the pgbase binary is executable
RUN chmod +x pgbase

# Expose the port that pgbase will run on
EXPOSE 8090

# Run the pgbase serve command
CMD ["./pgbase", "serve", "--http=0.0.0.0:8090"]
