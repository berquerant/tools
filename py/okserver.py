from http.server import (
    BaseHTTPRequestHandler,
    HTTPServer,
)
from http import HTTPStatus
import ssl
from uuid import uuid4
import json


DEFAULT_BUFFER_SIZE_BYTE = 1000


class Handler(BaseHTTPRequestHandler):

    def write_log(self, req_id: str):
        cl = int(self.headers.get('content-length', DEFAULT_BUFFER_SIZE_BYTE))
        ld = [
            '[{}]'.format(req_id),
            self.command,
            self.path,
            self.request_version,
            json.dumps(dict(self.headers.items())),
        ]
        if self.command == 'POST':
            ld.append(self.rfile.read(cl).decode('utf-8'))
        self.log_message(' '.join(ld))

    def gen_response_body(self, req_id: str) -> str:
        return json.dumps({
            'req_id': req_id,
            'message': 'OK',
        })

    def gen_req_id(self) -> str:
        return str(uuid4())

    def handle_request(self):
        req_id = self.gen_req_id()
        self.write_log(req_id)
        self.send_response(HTTPStatus.OK)
        self.send_header('content-type', 'application/json')
        self.end_headers()
        self.wfile.write(self.gen_response_body(req_id).encode('utf-8'))

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


if __name__ == '__main__':
    from argparse import ArgumentParser

    p = ArgumentParser(
        description='HTTP(S) server that responds 200 OK',
    )
    p.add_argument('--host', action='store', type=str, default='localhost', help='host')
    p.add_argument('-p', action='store', type=int, default=8080, help='port')
    p.add_argument('-c', action='store', type=str, help='certification')
    p.add_argument('-k', action='store', type=str, help='secret key')
    opt = p.parse_args()

    s = Server(host=opt.host, port=opt.p, handler=Handler)
    ctx = Server.new_context(crt=opt.c, key=opt.k) if opt.c and opt.k else None
    s.run(ctx=ctx)
