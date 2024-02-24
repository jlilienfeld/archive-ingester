
FROM golang:1.22-bookworm as builder

# Create and change to the app directory.
WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -v -o archive-ingester

FROM debian:bookworm-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/archive-ingester /app/archive-ingester
COPY assets/index.html /app/assets/index.html

ENV repopath=/data

WORKDIR /app

# Run the web service on container startup.
CMD ["/app/archive-ingester"]

