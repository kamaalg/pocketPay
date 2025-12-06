FROM golang:1.25.4-alpine
WORKDIR /app
COPY . .
WORKDIR /app/auth_service
COPY ./auth_service/go.mod ./auth_service/go.sum ./

RUN go mod download

COPY . .

RUN go build -o server ./cmd/api



CMD ["./server"]

