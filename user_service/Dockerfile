FROM golang:1.25.4-alpine
WORKDIR /app
COPY . .
WORKDIR /app/user_service
COPY ./user_service/go.mod ./user_service/go.sum ./

RUN go mod download

COPY . .

RUN go build -o server ./cmd/api

EXPOSE 8000


CMD ["./server"]

