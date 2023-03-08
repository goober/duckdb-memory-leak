FROM golang:1.20 as builder

RUN apt-get update \
    && apt-get install -y --no-install-recommends curl unzip \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN curl -sL https://github.com/duckdb/duckdb/releases/download/v0.7.1/libduckdb-linux-amd64.zip -o /tmp/libduckdb.zip; \
    unzip /tmp/libduckdb.zip -d /opt/duckdb

COPY ./go.mod /app/
COPY ./go.sum /app/

RUN go mod download

COPY . /app

RUN CGO_ENABLED=1 CGO_LDFLAGS="-L/opt/duckdb/" go build -tags=duckdb_use_lib -o dist/duckdb-troubleshooting ./...

FROM debian:11-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /opt/duckdb /opt/duckdb
COPY --from=builder /app/dist/duckdb-troubleshooting /
ENV LD_LIBRARY_PATH=/opt/duckdb
ENTRYPOINT ["/duckdb-troubleshooting"]
