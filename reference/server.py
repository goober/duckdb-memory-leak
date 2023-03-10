from http.server import BaseHTTPRequestHandler, HTTPServer
from prometheus_client import start_http_server
import duckdb
import json


class ServerHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        conn = duckdb.connect(database="", read_only=False)
        total_query = conn.sql("select count(*) as total from 'data.parquet'")
        total = total_query.fetchone()[0]
        total_query.close()
        counter = 0
        while counter < total:
            query = conn.sql("select total_amount from 'data.parquet' LIMIT 50000 OFFSET {}".format(counter))
            rows = query.fetchall()
            for row in rows:
                counter = counter + 1
            query.close()

        conn.close()
        payload = {
            "total": counter
        }
        self.send_response(200)
        self.headers.add_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(bytes(json.dumps(payload), "utf-8"))


if __name__ == "__main__":
    start_http_server(8000)
    server = HTTPServer(('', 8080), ServerHandler)
    print("Server started http://localhost:8080")

    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass

    server.server_close()
    print("Server stopped.")
