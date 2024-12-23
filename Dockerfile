# Build
# ----------

FROM golang:1.23-bookworm AS builder

WORKDIR /app

# Cache dependencies
COPY go.* ./
RUN go mod download

COPY . ./
RUN go build -v -o service cmd/service/main.go

# Runtime
# ----------

FROM debian:bookworm-slim

# Install package dependencies and remove apt lists
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy agent binary
COPY --from=builder /app/agent /app/agent

# Execute agent
ENTRYPOINT ["/app/agent"]
