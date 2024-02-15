# Stage 1: Build the application
FROM golang:alpine AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the source code
COPY . .

# Build the Go application with additional flags to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o app

# Stage 2: Create the final lightweight image
FROM scratch

# Copy the built binary from the build stage
COPY --from=build /app/app /

# Set the command to run the executable
ENTRYPOINT ["/app"]
