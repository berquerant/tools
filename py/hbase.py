import happybase as hb
import json
from contextlib import contextmanager
import sys


ROWS_UNIT = 1000


def decode(b)-> str:
    try:
        return b.decode('utf-8')
    except Exception:
        return str(b)


class HBase:
    def __init__(self, host: str, port: int):
        self.host = host
        self.port = port

    @contextmanager
    def conn(self):
        c = hb.Connection(self.host, self.port)
        try:
            yield c
        finally:
            c.close()

    def create_table(self, schema: dict):
        with self.conn() as c:
            c.create_table(**schema)

    def truncate(self):
        with self.conn() as c:
            tables = c.tables()
            for t in tables:
                c.delete_table(t, disable=True)

    def list_tables(self)-> list:
        with self.conn() as c:
            return c.tables()

    def list_tables_and_dump(self):
        for t in self.list_tables():
            print(decode(t))

    def scan(self, table: str, prefix='')-> iter:
        with self.conn() as c:
            if prefix:
                return c.table(table).scan(row_prefix=prefix.encode('utf-8'))
            return c.table(table).scan()

    def scan_and_dump(self, table: str, prefix=''):
        for k, v in self.scan(table, prefix):
            print('{} {}'.format(
                decode(k),
                json.dumps(
                    {
                        decode(ik): decode(iv)
                    } for ik, iv in v.items()
                ),
            ))

    def rows(self, table: str, keys: list)-> list:
        with self.conn() as c:
            return c.table(table).rows(keys)

    def rows_and_dump(self, table: str, keys: list):
        for k, v in self.rows(table, keys):
            print('{} {}'.format(
                decode(k),
                json.dumps(
                    {
                        decode(ik): decode(iv)
                    } for ik, iv in v.items()
                ),
            ))


if __name__ == '__main__':
    from argparse import ArgumentParser

    p = ArgumentParser(
        description='manipulate hbase',
    )
    # common
    p.add_argument('--host', action='store', type=str, required=True, help='host')
    p.add_argument('--port', action='store', type=int, default=9090, help='port')
    sp = p.add_subparsers(dest='command')
    # create table
    sp_create = sp.add_parser('create_table', help='create table, read schema from stdin')
    # truncate
    sp_truncate = sp.add_parser('truncate', help='truncate tables')
    # list tables
    sp_list = sp.add_parser('list', help='list tables')
    # scan
    sp_scan = sp.add_parser('scan', help='scan table')
    sp_scan.add_argument('-t', action='store', type=str, required=True, help='table')
    sp_scan.add_argument('--prefix', action='store', type=str, default='', help='prefix')
    # rows
    sp_rows = sp.add_parser('rows', help='get rows, read keys from stdin')
    sp_rows.add_argument('-t', action='store', type=str, required=True, help='table')

    opt = p.parse_args()
    hbase = HBase(host=opt.host, port=opt.port)

    def main(h, cmd):
        if cmd == 'create_table':
            h.create_table(json.loads(sys.stdin.read()))
            return

        if cmd == 'truncate':
            h.truncate()
            return

        if cmd == 'list':
            h.list_tables_and_dump()
            return

        if cmd == 'scan':
            h.scan_and_dump(table=opt.t, prefix=opt.prefix)
            return

        if cmd == 'rows':
            keys = []
            for line in sys.stdin:
                x = line.rstrip()
                if not x:
                    continue
                keys.append(x)
                if len(keys) == ROWS_UNIT:
                    h.rows_and_dump(table=opt.t, keys=keys)
                    keys = []
            h.rows_and_dump(table=opt.t, keys=keys)
            return

    main(hbase, opt.command)
