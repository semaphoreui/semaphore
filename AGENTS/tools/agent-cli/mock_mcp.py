"""Fake MCP endpoint: logs method, path and Authorization header, answers 401."""
import http.server
import sys

port, log = int(sys.argv[1]), sys.argv[2]


class Handler(http.server.BaseHTTPRequestHandler):
    def _handle(self):
        with open(log, "a") as f:
            f.write(f"{self.command} {self.path} auth={self.headers.get('Authorization')}\n")
        self.send_response(401)
        self.send_header("Content-Length", "0")
        self.end_headers()

    do_GET = do_POST = do_DELETE = _handle

    def log_message(self, *args):
        pass


http.server.HTTPServer(("127.0.0.1", port), Handler).serve_forever()
