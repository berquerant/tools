import boto3
import json
import sys
from time import sleep
from common.json import Encoder


UPLOAD_INTERVAL_SEC = 10
UPLOAD_SIZE_BYTE = 5*1024*1024 - 500*1024

RETURN_ALL_FIELDS = '_all_fields'
RETURN_NO_FIELDS = '_no_fields'
RETURN_SCORE = '_score'

SIMPLE_PARSER = 'simple'
SIMPLE_PARSER_MATCH_ALL = '*'

SCORE_DESC = '_score desc'


def split_by_size(xs: iter, size_byte: int)-> iter:
    acc = []
    acc_byte = 0
    for x in xs:
        x_byte = len(json.dumps(x).encode('utf-8'))
        sum_byte = acc_byte + x_byte
        if sum_byte >= size_byte:
            yield acc
            acc = [x]
            acc_byte = x_byte
        else:
            acc.append(x)
            acc_byte = sum_byte
    if acc:
        yield acc


class Cloudsearch:
    def __init__(self, region_name: str, search_endpoint='', document_endpoint='', no_mercy=False):
        self.region_name = region_name
        self.search_endpoint = search_endpoint
        self.document_endpoint = document_endpoint
        self.no_mercy = no_mercy

    def client(self, endpoint_url: str):
        return boto3.client('cloudsearchdomain', region_name=self.region_name, endpoint_url=endpoint_url)

    def search_client(self):
        return self.client(self.search_endpoint)

    def document_client(self):
        return self.client(self.document_endpoint)

    def _sleep_interval(self):
        if not self.no_mercy:
            sleep(UPLOAD_INTERVAL_SEC)

    def _upload(self, documents: bytes):
        c = self.document_client()
        retry = 0
        while True:
            try:
                r = c.upload_documents(
                    documents=documents,
                    contentType='application/json',
                )
                return r
            except Exception as e:
                print(e, file=sys.stderr)
                if retry > 3:
                    raise Exception('retry exhasusted: {}'.format(e))
                self._sleep_interval()
                retry += 1

    def upload(self, documents: iter)-> iter:
        for doc in split_by_size(xs=documents, size_byte=UPLOAD_SIZE_BYTE):
            yield self._upload(json.dumps(doc).encode('utf-8'))
            self._sleep_interval()

    def upload_and_dump(self, documents: iter):
        for x in self.upload(documents):
            print(json.dumps(x, cls=Encoder))

    def truncate(self):
        retry = 0
        while True:
            try:
                ids = [
                    x['id'] for x in self.search(
                        query=SIMPLE_PARSER_MATCH_ALL,
                        parser=SIMPLE_PARSER,
                        fields=RETURN_NO_FIELDS,
                        sort=SCORE_DESC,
                        limit=1000,
                    )
                ]
                if not ids:
                    return
                self.upload_and_dump({
                    'type': 'delete',
                    'id': x,
                } for x in ids)
            except Exception as e:
                print(e, file=sys.stderr)
                if retry >= 3:
                    raise Exception('retry exhasusted: {}'.format(e))
                print('sleep {} sec'.format(2**retry))
                sleep(2**retry)
                retry += 1

    def search(self, query: str, parser: str, fields: str, sort: str, size=1000, limit=0)-> iter:
        c = self.search_client()

        def params(cursor: str)-> dict:
            return {
                'size': size,
                'sort': sort,
                'cursor': cursor,
                'query': query,
                'queryParser': parser,
                'returnFields': fields,
            }

        def do(cursor='initial', count=0):
            r = c.search(**params(cursor))
            if not r['hits']['hit']:
                return
            for h in r['hits']['hit']:
                if limit > 0 and count >= limit:
                    return
                yield {
                    k: v for k, v in h.items()
                    if k in ['id', 'fields']
                }
                count += 1
            if 'cursor' not in r['hits']:
                return
            yield from do(cursor=r['hits']['cursor'], count=count)

        yield from do()

    def search_and_dump(self, query: str, parser: str, fields: str, sort: str, size=1000):
        for x in self.search(query=query, parser=parser, fields=fields, sort=sort, size=size):
            print(json.dumps(x, cls=Encoder))


if __name__ == '__main__':
    from argparse import ArgumentParser

    p = ArgumentParser(
        description='manipulate cloudsearch',
    )

    # common
    p.add_argument('--region_name', action='store', type=str, required=True, help='region')
    p.add_argument('--no_mercy', action='store_const', const=True, default=False, help='no upload interval')
    sp = p.add_subparsers(dest='command')
    # search
    sp_search = sp.add_parser('search', help='search')
    sp_search.add_argument('--endpoint_url', action='store', type=str, required=True, help='search endpoint')
    sp_search.add_argument('-q', action='store', type=str, default=SIMPLE_PARSER_MATCH_ALL, help='query')
    sp_search.add_argument('-p', action='store', type=str, default=SIMPLE_PARSER, help='query parser')
    sp_search.add_argument('-f', action='store', default=RETURN_ALL_FIELDS,
                           help='fields to include in response, comma separated list, see cloudsearchdomain')
    sp_search.add_argument('--sort', action='store', default=SCORE_DESC,
                           help='sort search results, comma separated list, see cloudsearchdomain')
    # upload
    sp_upload = sp.add_parser('upload', help='upload, read documents from stdin, json a line')
    sp_upload.add_argument('--endpoint_url', action='store', type=str, required=True, help='document endpoint')
    # truncate
    sp_truncate = sp.add_parser('truncate', help='truncate')
    sp_truncate.add_argument('--search_endpoint', action='store', type=str, required=True, help='search endpoint')
    sp_truncate.add_argument('--document_endpoint', action='store', type=str, required=True, help='document endpoint')

    opt = p.parse_args()
    conf = {
        'region_name': opt.region_name,
        'no_mercy': opt.no_mercy,
    }
    if opt.command == 'search':
        conf['search_endpoint'] = opt.endpoint_url
    if opt.command == 'upload':
        conf['document_endpoint'] = opt.endpoint_url
    if opt.command == 'truncate':
        conf['search_endpoint'] = opt.search_endpoint
        conf['document_endpoint'] = opt.document_endpoint
    cloudsearch = Cloudsearch(**conf)

    def main(c, cmd):
        if cmd == 'search':
            c.search_and_dump(query=opt.q, parser=opt.p, fields=opt.f, sort=opt.sort)
            return

        if cmd == 'upload':
            c.upload_and_dump(json.loads(line.rstrip()) for line in sys.stdin)
            return

        if cmd == 'truncate':
            c.truncate()
            return

    main(cloudsearch, opt.command)
