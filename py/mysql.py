import json
from mysql import connector
from common.json import Encoder


class MySQL:
    def __init__(self, user: str, password: str, host: str, port: int, database: str):
        conn = connector.connect(
            user=user,
            password=password,
            host=host,
            port=port,
            database=database,
        )
        if not conn.is_connected():
            conn.ping(True)
        self.conn = conn

    def select(self, query: str)-> list:
        c = self.conn.cursor(dictionary=True)
        c.execute(query)
        return c.fetchall()

    def select_and_dump(self, query: str):
        print(json.dumps(self.select(query), cls=Encoder))

    def modify(self, query: str)-> str:
        c = self.conn.cursor()
        try:
            c.execute(query)
            self.conn.commit()
            return '{} row(s) affected'.format(c.rowcount)
        except Exception as e:
            self.conn.rollback()
            return str(e)

    def dispatch(self, query: str)-> str:
        if query.split()[0].lower() in ['insert', 'update', 'delete']:
            print(self.modify(query))
            return
        self.select(query)


if __name__ == '__main__':
    from argparse import ArgumentParser

    p = ArgumentParser(
        description='jsonify result of query',
    )

    p.add_argument('-u', action='store', type=str, required=True, help='user')
    p.add_argument('-p', action='store', type=str, required=True, help='password')
    p.add_argument('--host', action='store', type=str, required=True, help='host')
    p.add_argument('--port', action='store', type=int, default=3306, help='port')
    p.add_argument('-d', action='store', type=str, required=True, help='database')
    p.add_argument('-q', action='store', type=str, required=True, help='query')

    opt = p.parse_args()
    sql = MySQL(
        user=opt.u,
        password=opt.p,
        host=opt.host,
        port=opt.port,
        database=opt.d,
    )
    sql.dispatch(opt.q)
