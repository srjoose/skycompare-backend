
# Build stage
FROM golang:1.24.2 AS builder


WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . . 

RUN go build -o server ./main/server.go

# Run stage
FROM alpine:latest  

WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
