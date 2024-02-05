# Build stage
FROM golang:1.20

WORKDIR /app

COPY . .

RUN GOOS=linux go build -o server ./cmd/server

EXPOSE 8080

CMD ["./server"]
