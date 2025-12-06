FROM golang:1.25.4-alpine
WORKDIR /app
COPY . .
WORKDIR /app/payment_service
COPY ./payment_service/go.mod ./payment_service/go.sum ./

RUN go mod download

COPY . .

RUN go build -o server ./cmd/api

EXPOSE 8001

CMD [ "./server" ]