FROM golang:1.25.4-alpine
WORKDIR /app
COPY . .
WORKDIR /app/transaction_service
COPY ./transaction_service/go.mod ./transaction_service/go.sum ./

RUN go mod download

COPY . .

RUN go build -o server ./cmd/api

EXPOSE 8001

CMD [ "./server" ]