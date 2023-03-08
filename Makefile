.PHONY: download
download:
	curl -sL -o data.parquet https://d37ci6vzurychx.cloudfront.net/trip-data/yellow_tripdata_2022-01.parquet

.PHONY: build
build:
	CGO_ENABLED=1 go build -o dist/duckdb-troubleshooting ./...

.PHONY: run
run: build
	dist/duckdb-troubleshooting

.PHONY: clean
clean:
	rm -rf dist
