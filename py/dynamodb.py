import boto3
import json
import sys
from boto3.dynamodb.conditions import Key
from common.json import Encoder
from common.utils import split_list


BATCH_WRITE_SIZE = 25
BATCH_GET_SIZE = 100


def simplify_item(self, dynamo_item: dict)-> dict:
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
        self.name = name
        self.typ = typ

    def val_dynamo(self, v)-> dict:
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
    def __init__(self, region_name: str, endpoint_url: str, table: str, pk=None, sk=None):
        conf = {
            'region_name': region_name
        }
        if endpoint_url:
            conf['endpoint_url'] = endpoint_url
        self.resource = boto3.resource('dynamodb', **conf)
        self.client = boto3.client('dynamodb', **conf)
        self.table = table
        self.pk = pk
        self.sk = sk

    def batch_write(self, items: list, batch_size=BATCH_WRITE_SIZE):
        with self.resource.Table(self.table).batch_writer() as b:
            for xs in split_list(xs=items, n=batch_size):
                for x in xs:
                    b.put_item(Item=x)

    def batch_delete(self, keys: list, batch_size=BATCH_WRITE_SIZE):
        with self.resource.Table(self.table).batch_writer() as b:
            for xs in split_list(xs=keys, n=batch_size):
                for x in xs:
                    b.delete_item(Key={
                        p.name: p.val(x[p.name]) for p
                        in [self.pk, self.sk] if p
                    })

    def batch_get(self, keys: list, batch_size=BATCH_GET_SIZE):
        for xs in split_list(xs=keys, n=batch_size):
            yield from (
                simplify_item(y) for y
                in self.client.batch_get_item(
                    RequestItems={
                        self.table: {
                            'Keys': [
                                {
                                    p.name: p.val_dynamo(x[p.name]) for p
                                    in [self.pk, self.sk] if p
                                } for x in xs
                            ],
                        },
                    },
                )['Responses'][self.table]
            )

    def batch_get_and_dump(self, keys: list):
        for x in self.batch_get(keys=keys):
            print(json.dumps(x, cls=Encoder))

    def query(self, pv: str, limit=1000)-> iter:
        def params(last_key='')-> dict:
            p = {
                'Limit': limit,
                'KeyConditionExpression': Key(self.pk.name).eq(self.pk.val(pv)),
            }
            if last_key:
                p['ExclusiveStartKey'] = last_key
            return p

        def do(last_key=''):
            r = self.resource.Table(self.table).query(**params(last_key))
            yield from r['Items']
            lk = r.get('LastEvaluatedKey', '')
            if not lk:
                return
            yield from do(last_key=lk)

        yield from do()

    def query_and_dump(self, pv: str, limit=1000):
        for x in self.query(pv=pv, limit=limit):
            print(json.dumps(x, cls=Encoder))


if __name__ == '__main__':
    from argparse import ArgumentParser

    p = ArgumentParser(
        description='manipulate dynamodb',
    )
    # common
    p.add_argument('--region_name', action='store', type=str, required=True, help='region')
    p.add_argument('--endpoint_url', action='store', default='', help='endpoint')
    p.add_argument('-t', action='store', type=str, required=True, help='table')
    sp = p.add_subparsers(dest='command')
    # query
    sp_query = sp.add_parser('query', help='query')
    sp_query.add_argument('-pk', action='store', type=str, required=True, help='name of primary key')
    sp_query.add_argument('-pt', action='store', type=str, required=True, help='type of primary key')
    sp_query.add_argument('-pv', action='store', type=str, required=True, help='value of primary key')
    # get
    sp_get = sp.add_parser('get', help='batch get, read keys from stdin, json, {key_name: key_value}')
    sp_get.add_argument('-pk', action='store', type=str, required=True, help='name of primary key')
    sp_get.add_argument('-pt', action='store', type=str, required=True, help='type of primary key')
    sp_get.add_argument('-sk', action='store', type=str, default='', help='name of sort key')
    sp_get.add_argument('-st', action='store', type=str, default='', help='type of sort key')
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
    }
    if hasattr(opt, 'pk') and hasattr(opt, 'pt'):
        conf['pk'] = DynamoKey(name=opt.pk, typ=opt.pt)
    if hasattr(opt, 'sk') and hasattr(opt, 'st'):
        conf['sk'] = DynamoKey(name=opt.sk, typ=opt.st)
    dynamo = Dynamo(**conf)

    def main(d, cmd: str):
        if cmd == 'query':
            d.query_and_dump(pv=opt.pv)
            return

        def read_lines(size: int)-> iter:
            xs = []
            for line in sys.stdin:
                x = line.rstrip()
                if not x:
                    continue
                xs.append(x)
                if len(xs) == size:
                    yield xs
                    xs = []
            if xs:
                yield xs

        if cmd == 'get':
            for xs in read_lines(BATCH_GET_SIZE):
                d.batch_get_and_dump(keys=[
                    json.loads(x) for x in xs
                ])
            return

        if cmd == 'delete':
            for xs in read_lines(BATCH_WRITE_SIZE):
                d.batch_delete(keys=[
                    json.loads(x) for x in xs
                ])
            return

        if cmd == 'write':
            for xs in read_lines(BATCH_WRITE_SIZE):
                d.batch_write(items=[
                    json.loads(x) for x in xs
                ])
            return

    main(dynamo, opt.command)
