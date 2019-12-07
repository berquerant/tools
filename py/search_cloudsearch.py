import boto3


class Cloudsearch:
    def __init__(self, region_name: str, endpoint_url=''):
        conf = {
            'region_name': region_name,
        }
        if endpoint_url:
            conf['endpoint_url'] = endpoint_url
        self.client = boto3.client('cloudsearchdomain', **conf)

    def search(self, query: str, parser: str, fields: str, size=1000)-> iter:
        def params(cursor: str)-> dict:
            return {
                'size': size,
                'cursor': cursor,
                'query': query,
                'queryParser': parser,
                'returnFields': fields,
            }

        def do(cursor='initial'):
            r = self.client.search(**params(cursor))
            if not r['hits']['hit']:
                return
            for h in r['hits']['hit']:
                yield {
                    k: v for k, v in h.items()
                    if k in ['id', 'fields']
                }
            if 'cursor' not in r['hits']:
                return
            yield from do(cursor=r['hits']['cursor'])

        yield from do()


if __name__ == '__main__':
    from argparse import ArgumentParser

    p = ArgumentParser(
        description='search from cloudsearch',
    )

    p.add_argument('--region_name', action='store', type=str, required=True, help='region')
    p.add_argument('--endpoint_url', action='store', default='', help='endpoint')
    p.add_argument('--q', action='store', type=str, required=True, help='query')
    p.add_argument('--p', action='store', type=str, required=True, help='query parser')
    p.add_argument('--f', action='store', default='_all_fields', help='fields to include in response, comma separated list, see cloudsearchdomain')

    opt = p.parse_args()
    cloudsearch = Cloudsearch(
        region_name=opt.region_name,
        endpoint_url=opt.endpoint_url,
    )
    for x in cloudsearch.search(query=opt.q, parser=opt.p, fields=opt.f):
        print(x)
