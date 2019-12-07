import boto3
import json
from boto3.dynamodb.conditions import Key
from common.json import Encoder


class Dynamo:
    def __init__(self, region_name: str, endpoint_url: str):
        conf = {
            'region_name': region_name
        }
        if endpoint_url:
            conf['endpoint_url'] = endpoint_url
        self.resource = boto3.resource('dynamodb', **conf)

    def query(self, table: str, pk: str, pt: str, pv: str, limit=1000)-> iter:
        def conv():
            if pt in ['N', 'BOOL']:
                return int(pv)
            return pv

        pv = conv()

        def params(last_key='')-> dict:
            p = {
                'Limit': limit,
                'KeyConditionExpression': Key(pk).eq(pv),
            }
            if last_key:
                p['ExclusiveStartKey'] = last_key
            return p

        def do(last_key=''):
            r = self.resource.Table(table).query(**params(last_key))
            yield from r['Items']
            lk = r.get('LastEvaluatedKey', '')
            if not lk:
                return
            yield from do(last_key=lk)

        yield from do()

    def query_and_dump(self, table: str, pk: str, pt: str, pv: str, limit=1000):
        for x in self.query(table=table, pk=pk, pt=pt, pv=pv, limit=limit):
            print(json.dumps(x, cls=Encoder))


if __name__ == '__main__':
    from argparse import ArgumentParser

    p = ArgumentParser(
        description='query from dynamodb',
    )

    p.add_argument('--region_name', action='store', type=str, required=True, help='region')
    p.add_argument('--endpoint_url', action='store', default='', help='endpoint')
    p.add_argument('-t', action='store', type=str, required=True, help='table')
    p.add_argument('-pk', action='store', type=str, required=True, help='name of primary key')
    p.add_argument('-pt', action='store', type=str, required=True, help='type of primary key')
    p.add_argument('-pv', action='store', type=str, required=True, help='value of primary key')

    opt = p.parse_args()
    dynamo = Dynamo(
        region_name=opt.region_name,
        endpoint_url=opt.endpoint_url,
    )
    dynamo.query_and_dump(
        table=opt.t,
        pk=opt.pk,
        pt=opt.pt,
        pv=opt.pv,
    )
