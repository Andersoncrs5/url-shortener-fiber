FROM golang:1.25.4 AS builder 
WORKDIR /app

RUN apt-get update && apt-get install -y \
    build-essential \
    librdkafka-dev \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o write_api main.go

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

RUN apt-get update && apt-get install -y \
    librdkafka1 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/write_api /write_api

EXPOSE 8080

ENTRYPOINT ["/write_api"]