#!/usr/bin/env python3
"""Minimal HTTP server wrapping argos-translate.

Usage:
  python3 argos_server.py [--port 5000]

API:
  POST /translate
  {"q": "text", "source": "en", "target": "it"}
  → {"translatedText": "testo"}
"""

import json
import sys
from http.server import HTTPServer, BaseHTTPRequestHandler

import argostranslate.translate


class TranslateHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(length)
        try:
            req = json.loads(body)
            text = req.get("q", req.get("text", ""))
            source = req.get("source", "en")
            target = req.get("target", "it")

            result = argostranslate.translate.translate(text, source, target)

            resp = json.dumps({"translatedText": result}).encode()
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(resp)))
            self.end_headers()
            self.wfile.write(resp)
        except Exception as e:
            resp = json.dumps({"error": str(e)}).encode()
            self.send_response(500)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(resp)))
            self.end_headers()
            self.wfile.write(resp)

    def log_message(self, fmt, *args):
        sys.stderr.write(f"[argos] {fmt % args}\n")


if __name__ == "__main__":
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 5000
    server = HTTPServer(("127.0.0.1", port), TranslateHandler)
    print(f"argos-translate server on http://127.0.0.1:{port}", file=sys.stderr)
    server.serve_forever()
