from http.server import (
    BaseHTTPRequestHandler,
    HTTPServer,
)
from http import HTTPStatus
import ssl
from uuid import uuid4
import json
import requests
from urllib.parse import (
    urlparse,
    urlunparse,
)
import time


BUFFER_SIZE_BYTE = 1000
DESTINATION = ""
USE_PROPAGATE_PATH = False
USE_LIGHT_LOG = False
FORWARD_DELAY_SEC = 0
FORWARDED_TIMEOUT_SEC = 0
RESPONSE_DELAY_SEC = 0


class Handler(BaseHTTPRequestHandler):
    def get_content_length(self) -> int:
        return int(self.headers.get("content-length", BUFFER_SIZE_BYTE))

    def write_log(self, req_id: str) -> str:
        cl = self.get_content_length()
        ld = [
            "[{}]".format(req_id),
            self.command,
            self.path,
            self.request_version,
            json.dumps(dict(self.headers.items()), separators=(",", ":")),
        ]
        if self.command == "GET":
            self.log_message(" ".join(ld))
        data = self.rfile.read(cl).decode("utf-8")
        ld.append(data)
        self.log_message(" ".join(ld))
        return data

    @staticmethod
    def gen_response_body(req_id: str) -> str:
        return json.dumps(
            {
                "req_id": req_id,
                "message": "OK",
            }
        )

    @staticmethod
    def gen_req_id() -> str:
        return str(uuid4())

    @staticmethod
    def gen_destination(url: str) -> str:
        if not USE_PROPAGATE_PATH:
            return DESTINATION
        u = urlparse(url)
        ep = urlparse(DESTINATION + u.path)
        return urlunparse(
            (ep.scheme, ep.netloc, ep.path, u.params, u.query, u.fragment)
        )

    def forward_request(self, req_id: str, data: str = None):
        url = self.gen_destination(self.path)
        self.log_message(
            "[{}] forward to {} after {} sec".format(req_id, url, FORWARD_DELAY_SEC)
        )
        time.sleep(FORWARD_DELAY_SEC)
        hs = dict(self.headers.items())
        hs["x-forward-request-id"] = req_id
        args = {
            "url": url,
            "headers": hs,
            "timeout": FORWARDED_TIMEOUT_SEC,
        }
        if self.command == "GET":
            return requests.get(**args)
        args["data"] = data
        return requests.post(**args)

    def handle_request(self):
        req_id = self.gen_req_id()
        data = self.write_log(req_id)
        if not DESTINATION:
            self.log_message(
                "[{}] response after {} sec".format(req_id, RESPONSE_DELAY_SEC)
            )
            time.sleep(RESPONSE_DELAY_SEC)
            self.send_response(HTTPStatus.OK)
            self.send_header("content-type", "application/json")
            self.end_headers()
            self.wfile.write(self.gen_response_body(req_id).encode("utf-8"))
            return

        r = self.forward_request(req_id, data)
        if not USE_LIGHT_LOG:
            self.log_message(
                " ".join(
                    [
                        "[{}]".format(req_id),
                        "forwarded",
                        str(r.status_code),
                        json.dumps(dict(r.headers), separators=(",", ":")),
                        r.text,
                    ]
                )
            )
        self.log_message(
            "[{}] response after {} sec".format(req_id, RESPONSE_DELAY_SEC)
        )
        time.sleep(RESPONSE_DELAY_SEC)
        self.send_response(r.status_code)
        for k, v in r.headers.items():
            self.send_header(k, v)
        self.end_headers()
        self.wfile.write(r.text.encode("utf-8"))

    def do_GET(self):
        self.handle_request()

    def do_POST(self):
        self.handle_request()


class Server:
    def __init__(self, host: str, port: int, handler):
        self.host = host
        self.port = port
        self.handler = handler

    @staticmethod
    def new_context(crt: str, key: str):
        ctx = ssl.create_default_context(ssl.Purpose.CLIENT_AUTH)
        ctx.load_cert_chain(crt, keyfile=key)
        ctx.options |= ssl.OP_NO_TLSv1 | ssl.OP_NO_TLSv1_1
        return ctx

    def run(self, ctx=None):
        s = HTTPServer((self.host, self.port), self.handler)
        if ctx:
            s.socket = ctx.wrap_socket(s.socket)
        try:
            s.serve_forever()
        except KeyboardInterrupt:
            pass
        s.server_close()


if __name__ == "__main__":
    from argparse import ArgumentParser

    p = ArgumentParser(
        description="HTTP(S) server that responds 200 OK",
    )
    p.add_argument("--host", action="store", type=str, default="localhost", help="host")
    p.add_argument("-p", action="store", type=int, default=8080, help="port")
    p.add_argument("-c", action="store", type=str, help="certification")
    p.add_argument("-k", action="store", type=str, help="secret key")
    p.add_argument(
        "-f", action="store", type=str, help="destination endpoint for forwarding"
    )
    p.add_argument(
        "-fd", action="store", type=float, default=0, help="delay second for forwarding"
    )
    p.add_argument(
        "-ft",
        action="store",
        type=float,
        default=60,
        help="timeout second for forwarded request",
    )
    p.add_argument("-s", action="store_true", help="propagate path")
    p.add_argument(
        "-b", action="store", type=int, default=1024, help="buffer size byte"
    )
    p.add_argument("-l", action="store_true", help="light log")
    p.add_argument(
        "-d", action="store", type=float, default=0, help="delay second to response"
    )
    opt = p.parse_args()

    s = Server(host=opt.host, port=opt.p, handler=Handler)
    ctx = Server.new_context(crt=opt.c, key=opt.k) if opt.c and opt.k else None
    DESTINATION = opt.f
    USE_PROPAGATE_PATH = opt.s
    USE_LIGHT_LOG = opt.l
    BUFFER_SIZE_BYTE = opt.b
    FORWARD_DELAY_SEC = opt.fd
    FORWARDED_TIMEOUT_SEC = opt.ft
    RESPONSE_DELAY_SEC = opt.d
    s.run(ctx=ctx)
