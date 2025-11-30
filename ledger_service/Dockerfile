## Builder stage: generate protobuf and build binary
FROM golang:1.20-bullseye AS builder

WORKDIR /workspace

RUN apt-get update && apt-get install -y --no-install-recommends protobuf-compiler git && \
    rm -rf /var/lib/apt/lists/*

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

COPY ledger.proto ./
COPY go.mod go.sum ./
RUN go mod download

RUN protoc --go_out=. --go-grpc_out=. ./ledger.proto

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /ledger_server ./cmd/server

FROM gcr.io/distroless/base-debian11
COPY --from=builder /ledger_server /ledger_server
EXPOSE 50051
ENTRYPOINT ["/ledger_server"]
