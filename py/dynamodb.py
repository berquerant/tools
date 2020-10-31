import boto3
import json
import sys
from boto3.dynamodb.conditions import Key
from common.json import dumps_compact
from common.utils import split_list
import time


BATCH_WRITE_SIZE = 25
BATCH_GET_SIZE = 100


def simplify_item(dynamo_item: dict) -> dict:
    def translate(s):
        k, v = list(s.keys())[0], list(s.values())[0]
        if k == 'N':
            x = float(v)
            if x.is_integer():
                return int(x)
            return x
        if k == 'M':
            return {
                a: translate(b) for a, b in v.items()
            }
        if k == 'L':
            return [
                translate(a) for a in v
            ]
        return v
    return {
        k: translate(v) for k, v in dynamo_item.items()
    }


class DynamoKey:
    def __init__(self, name: str, typ: str):
        self.__name = name
        self.__typ = typ

    @property
    def name(self) -> str:
        return self.__name

    @property
    def typ(self) -> str:
        return self.__typ

    def val_dynamo(self, v) -> dict:
        return {self.typ: v}

    def val(self, v):
        if not isinstance(v, str):
            return v
        if self.typ == 'BOOL':
            return v == 'true'
        if self.typ == 'N':
            x = float(v)
            if x.is_integer():
                return int(x)
            return x
        return v


class Dynamo:
    def __init__(self, region_name: str, table: str, endpoint_url='',
                 pk=None, sk=None, attrs: list = None, head=0, mercy_sec=0):
        conf = {
            'region_name': region_name,
        }
        if endpoint_url:
            conf['endpoint_url'] = endpoint_url
        self.__resource = boto3.resource('dynamodb', **conf)
        self.__client = boto3.client('dynamodb', **conf)
        self.__table = table
        self.__pk = pk
        self.__sk = sk
        self.__attrs = attrs
        self.__head = head
        self.__mercy_sec = mercy_sec

    @property
    def resource(self):
        return self.__resource

    @property
    def client(self):
        return self.__client

    @property
    def table(self) -> str:
        return self.__table

    @property
    def pk(self) -> str:
        return self.__pk

    @property
    def sk(self) -> str:
        return self.__sk

    @property
    def attrs(self) -> list:
        return self.__attrs

    @property
    def head(self) -> int:
        return self.__head

    def _projection_expression(self) -> str:
        if not self.attrs:
            return ''
        return ','.join(self.attrs)

    def _sleep_for_mercy(self):
        time.sleep(self.__mercy_sec)

    def batch_write(self, items: list, batch_size=BATCH_WRITE_SIZE):
        at_start = True
        with self.resource.Table(self.table).batch_writer() as b:
            for xs in split_list(xs=items, n=batch_size):
                if at_start:
                    at_start = False
                else:
                    self._sleep_for_mercy()
                for x in xs:
                    b.put_item(Item=x)

    def batch_delete(self, keys: list, batch_size=BATCH_WRITE_SIZE):
        at_start = True
        with self.resource.Table(self.table).batch_writer() as b:
            for xs in split_list(xs=keys, n=batch_size):
                if at_start:
                    at_start = False
                else:
                    self._sleep_for_mercy()
                for x in xs:
                    b.delete_item(Key={
                        p.name: p.val(x[p.name]) for p
                        in [self.pk, self.sk] if p
                    })

    def batch_get(self, keys: list, batch_size=BATCH_GET_SIZE):
        at_start = True
        projection_expression = self._projection_expression()
        for xs in split_list(xs=keys, n=batch_size):
            if at_start:
                at_start = False
            else:
                self._sleep_for_mercy()
            items = {
                'Keys': [
                    {
                        p.name: p.val_dynamo(x[p.name]) for p
                        in [self.pk, self.sk] if p
                    } for x in xs
                ],
            }
            if projection_expression:
                items['ProjectionExpression'] = projection_expression
            yield from (
                simplify_item(y) for y
                in self.client.batch_get_item(
                    RequestItems={
                        self.table: items,
                    },
                )['Responses'][self.table]
            )

    def query(self, pv: str, limit=1000) -> iter:
        key_condition_expression = Key(self.pk.name).eq(self.pk.val(pv))
        projection_expression = self._projection_expression()
        t = self.resource.Table(self.table)

        def params(last_key='') -> dict:
            p = {
                'Limit': limit,
                'KeyConditionExpression': key_condition_expression,
            }
            if projection_expression:
                p['ProjectionExpression'] = projection_expression
            if last_key:
                p['ExclusiveStartKey'] = last_key
            return p

        def do(last_key='', count=0):
            if self.head and count >= self.head:
                return
            if last_key:
                self._sleep_for_mercy()
            r = t.query(**params(last_key))
            for x in r['Items']:
                if self.head and count >= self.head:
                    return
                count += 1
                yield x
            lk = r.get('LastEvaluatedKey', '')
            if not lk:
                return
            yield from do(last_key=lk, count=count)

        yield from do()

    def scan(self, limit=1000) -> iter:
        projection_expression = self._projection_expression()
        t = self.resource.Table(self.table)

        def params(last_key='') -> dict:
            p = {
                'Limit': limit,
            }
            if projection_expression:
                p['ProjectionExpression'] = projection_expression
            if last_key:
                p['ExclusiveStartKey'] = last_key
            return p

        def do(last_key='', count=0):
            if self.head and count >= self.head:
                return
            if last_key:
                self._sleep_for_mercy()
            r = t.scan(**params(last_key))
            for x in r['Items']:
                if self.head and count >= self.head:
                    return
                count += 1
                yield x
            lk = r.get('LastEvaluatedKey', '')
            if not lk:
                return
            yield from do(lk, count)

        yield from do()


def main(d: Dynamo, cmd: str):
    if cmd == 'scan':
        for x in d.scan():
            print(dumps_compact(x))
        return

    if cmd == 'query':
        for x in d.query(pv=opt.pv):
            print(dumps_compact(x))
        return

    def read_lines(size: int) -> iter:
        xs = []
        for line in sys.stdin:
            x = line.rstrip()
            if not x:
                continue
            xs.append(json.loads(x))
            if len(xs) == size:
                yield xs
                xs = []
        if xs:
            yield xs

    if cmd == 'get':
        for xs in read_lines(BATCH_GET_SIZE):
            for x in d.batch_get(keys=xs):
                print(dumps_compact(x))
        return

    if cmd == 'delete':
        for xs in read_lines(BATCH_WRITE_SIZE):
            d.batch_delete(keys=xs)
            for x in xs:
                print(dumps_compact(x))
        return

    if cmd == 'write':
        for xs in read_lines(BATCH_WRITE_SIZE):
            d.batch_write(items=xs)
            for x in xs:
                print(dumps_compact(x))
        return
    raise Exception('Unknown operation {}'.format(cmd))


if __name__ == '__main__':
    from argparse import ArgumentParser

    p = ArgumentParser(
        description='manipulate dynamodb using json',
    )
    # common
    p.add_argument('--region_name', action='store', type=str, required=True, help='region')
    p.add_argument('--endpoint_url', action='store', default='', help='endpoint')
    p.add_argument('-t', action='store', type=str, required=True, help='table')
    p.add_argument('--mercy_sec', action='store', type=float, default=0.5, help='mercy sleep second')
    sp = p.add_subparsers(dest='command')
    # scan
    sp_scan = sp.add_parser('scan', help='scan')
    sp_scan.add_argument('--head', action='store', type=int, default=0, help='max number of items to read, 0 means unlimited')
    sp_scan.add_argument('attrs', metavar='ATTR', type=str, nargs='*', help='attributes to get')
    # query
    sp_query = sp.add_parser('query', help='query')
    sp_query.add_argument('-pk', action='store', type=str, required=True, help='name of primary key')
    sp_query.add_argument('-pt', action='store', type=str, required=True, help='type of primary key')
    sp_query.add_argument('-pv', action='store', type=str, required=True, help='value of primary key')
    sp_query.add_argument('--head', action='store', type=int, default=0, help='max number of items to read, 0 means unlimited')
    sp_query.add_argument('attrs', metavar='ATTR', type=str, nargs='*', help='attributes to get')
    # get
    sp_get = sp.add_parser('get', help='batch get, read keys from stdin, json, {key_name: key_value}')
    sp_get.add_argument('-pk', action='store', type=str, required=True, help='name of primary key')
    sp_get.add_argument('-pt', action='store', type=str, required=True, help='type of primary key')
    sp_get.add_argument('-sk', action='store', type=str, default='', help='name of sort key')
    sp_get.add_argument('-st', action='store', type=str, default='', help='type of sort key')
    sp_get.add_argument('attrs', metavar='ATTR', type=str, nargs='*', help='attributes to get')
    # delete
    sp_delete = sp.add_parser('delete', help='batch delete, read sort keys from stdin, json, {key_name: key_value}')
    sp_delete.add_argument('-pk', action='store', type=str, required=True, help='name of primary key')
    sp_delete.add_argument('-pt', action='store', type=str, required=True, help='type of primary key')
    sp_delete.add_argument('-sk', action='store', type=str, default='', help='name of sort key')
    sp_delete.add_argument('-st', action='store', type=str, default='', help='type of sort key')
    # write
    sp_write = sp.add_parser('write', help='batch write, read items from stdin, json')

    opt = p.parse_args()
    conf = {
        'region_name': opt.region_name,
        'endpoint_url': opt.endpoint_url,
        'table': opt.t,
        'mercy_sec': opt.mercy_sec,
    }
    if hasattr(opt, 'pk') and opt.pk and hasattr(opt, 'pt') and opt.pt:
        conf['pk'] = DynamoKey(name=opt.pk, typ=opt.pt)
    if hasattr(opt, 'sk') and opt.sk and hasattr(opt, 'st') and opt.st:
        conf['sk'] = DynamoKey(name=opt.sk, typ=opt.st)
    if hasattr(opt, 'attrs') and opt.attrs:
        conf['attrs'] = opt.attrs
    if hasattr(opt, 'head') and opt.head:
        conf['head'] = opt.head
    dynamo = Dynamo(**conf)
    main(dynamo, opt.command)
