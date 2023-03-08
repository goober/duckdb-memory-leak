package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/marcboeker/go-duckdb"
	_ "github.com/marcboeker/go-duckdb"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
)

var source = "data.parquet"

func query(w http.ResponseWriter, req *http.Request) {
	db := newDBConnection()
	defer db.Close()
	conn, err := db.Conn(req.Context())
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(500)
		return
	}
	defer conn.Close()

	var counter int64
	var total int64

	// Get total number of rows
	res := conn.QueryRowContext(req.Context(), fmt.Sprintf("select count(*) from '%s'", source))
	err = res.Scan(&total)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	// Process the file in batches and increase the total counter for each processed row
	for counter < total {
		part, partErr := queryFile(req.Context(), conn, source, counter)
		if partErr != nil {
			panic(partErr)
		}
		counter += part
	}
	w.Write([]byte(fmt.Sprintf(`{"total": %d}`, counter)))
}

func queryFile(ctx context.Context, conn *sql.Conn, source string, offset int64) (int64, error) {
	var counter int64
	rows, err := conn.QueryContext(ctx, fmt.Sprintf("select total_amount from '%s' LIMIT 50000 OFFSET %d", source, offset))
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var dummy interface{}

	for rows.Next() {
		scanErr := rows.Scan(&dummy)
		if scanErr != nil {
			panic(scanErr)
		}
		counter++
	}

	if rows.Err() != nil {
		return 0, rows.Err()
	}
	err = rows.Close()
	if err != nil {
		return 0, err
	}
	return counter, nil
}

func newDBConnection() *sql.DB {
	connector, err := duckdb.NewConnector("", func(execer driver.ExecerContext) error {
		bootQueries := []string{
			"INSTALL 'parquet'",
			"LOAD 'parquet'",
			"INSTALL 'httpfs'",
			"LOAD 'httpfs'",
		}

		for _, qry := range bootQueries {
			_, err := execer.ExecContext(context.Background(), qry, nil)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(1)
	return db
}

func main() {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/query", query)

	fmt.Println("Server started on :8090")
	fmt.Println("Make a request to http://localhost:8090/query and monitor the application's memory usage for each request")
	http.ListenAndServe(":8090", nil)
}

func init() {
	_, err := os.Stat(source)
	if os.IsNotExist(err) {
		panic(fmt.Sprintf("Expected dataset not found.\nHave you downloaded the dataset with `make download`?\n"))
	}
}
