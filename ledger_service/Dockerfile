# ---- Build stage ----
FROM golang:1.25.4 AS builder
ARG TARGETOS=linux
ARG TARGETARCH=arm64

WORKDIR /workspace

COPY ledger_service/go.mod ledger_service/go.sum ./ledger_service/
COPY db /workspace/db

WORKDIR /workspace/ledger_service
RUN go mod download

RUN apt-get update && apt-get install -y --no-install-recommends \
      protobuf-compiler git \
    && rm -rf /var/lib/apt/lists/*

ENV PATH="/go/bin:${PATH}"
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

COPY ledger_service/ ./

WORKDIR /workspace/ledger_service
RUN ls -la /go/bin || true
RUN protoc \
    --plugin=/go/bin/protoc-gen-go --go_out=. \
    --plugin=/go/bin/protoc-gen-go-grpc --go-grpc_out=. \
    ledger.proto

RUN go mod tidy

# Build for the target OS/ARCH provided by buildkit (TARGETOS/TARGETARCH).
# This prevents exec format errors when building on Apple Silicon or other architectures.
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /ledger_server ./cmd/server
RUN chmod +x /ledger_server || true

# ---- Runtime stage ----
FROM gcr.io/distroless/base-debian11

COPY --from=builder /ledger_server /ledger_server
EXPOSE 50051
ENTRYPOINT ["/ledger_server"]
