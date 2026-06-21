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

import argostranslate.package
import argostranslate.translate


def download_language_pack(from_code, to_code):
    """Download and install language package if missing."""
    trans = argostranslate.translate.get_translation_from_codes(from_code, to_code)
    if trans is not None:
        return trans

    argostranslate.package.update_package_index()
    for pkg in argostranslate.package.get_available_packages():
        if pkg.from_code == from_code and pkg.to_code == to_code:
            print(f"[argos] downloading {from_code}→{to_code} package...", file=sys.stderr)
            sys.stderr.flush()
            install_path = pkg.download()
            argostranslate.package.install_from_path(install_path)
            return argostranslate.translate.get_translation_from_codes(from_code, to_code)

    return None


class TranslateHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(length)
        try:
            req = json.loads(body)
            text = req.get("q", req.get("text", ""))
            source = req.get("source", "en")
            target = req.get("target", "it")

            trans = download_language_pack(source, target)
            if trans is None:
                raise ValueError(
                    f"No language package available for {source}→{target}"
                )
            result = trans.translate(text)

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
