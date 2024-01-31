# syntax=docker/dockerfile:1

FROM golang:latest

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o /app/main cmd/runServer/main.go
EXPOSE 8080
CMD ["/app/main"]
