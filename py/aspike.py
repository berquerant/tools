import aerospike
from aerospike import exception as ex
import sys
import json
from contextlib import contextmanager
from common.json import Encoder


PK = 'pk'
MD = 'md'


class Aerospike:
    def __init__(self, host: str, port: int, namespace: str, setname: str):
        conf = {
            'hosts': [(host, port)],
        }
        self.namespace = namespace
        self.setname = setname
        self.conf = conf

    @contextmanager
    def conn(self):
        c = aerospike.client(self.conf).connect()
        try:
            yield c
        finally:
            c.close()

    def put(self, items: list):
        with self.conn() as c:
            for item in items:
                if PK not in item:
                    continue
                pk = item[PK]
                del item[PK]
                metadata = item.get(MD)
                if metadata:
                    del item[MD]
                c.put((self.namespace, self.setname, pk), item, metadata)

    def remove(self, keys: iter):
        with self.conn() as c:
            for k in keys:
                try:
                    c.remove((self.namespace, self.setname, k))
                except ex.RecordNotFound:
                    pass

    def get(self, keys: iter)-> iter:
        with self.conn() as c:
            for k in keys:
                try:
                    yield c.get((self.namespace, self.setname, k))
                except ex.RecordNotFound:
                    pass

    def get_and_dump(self, keys: iter):
        for x in self.get(keys):
            self.default_callback(x)

    def scan(self, callback: callable):
        with self.conn() as c:
            s = c.scan(self.namespace, self.setname)
            s.foreach(callback)

    @staticmethod
    def default_callback(input_tuple: tuple):
        _, md, rec = input_tuple
        print(json.dumps([rec, md], cls=Encoder))

    def scan_and_dump(self):
        self.scan(callback=self.default_callback)


if __name__ == '__main__':
    from argparse import ArgumentParser

    p = ArgumentParser(
        description='manipulate aerospike',
    )
    # common
    p.add_argument('--host', action='store', type=str, required=True)
    p.add_argument('--port', action='store', type=int, required=True)
    p.add_argument('-n', action='store', type=str, required=True, help='namespace')
    p.add_argument('-s', action='store', type=str, required=True, help='set')
    sp = p.add_subparsers(dest='command')
    # scan
    sp_scan = sp.add_parser('scan', help='scan')
    # get
    sp_get = sp.add_parser('get', help='get')
    # remove
    sp_remove = sp.add_parser('remove', help='remove')
    # put
    sp_put = sp.add_parser('put', help='put, {} for key, {} for metadata'.format(PK, MD))

    opt = p.parse_args()
    conf = {
        'host': opt.host,
        'port': opt.port,
        'namespace': opt.n,
        'setname': opt.s,
    }
    aspike = Aerospike(**conf)

    def main(a, cmd: str):
        if cmd == 'scan':
            a.scan_and_dump()
            return

        if cmd == 'get':
            a.get_and_dump(x.rstrip() for x in sys.stdin)
            return

        if cmd == 'remove':
            a.remove(x.rstrip() for x in sys.stdin)
            return

        if cmd == 'put':
            a.put(json.loads(x.rstrip()) for x in sys.stdin)
            return

    main(aspike, opt.command)
