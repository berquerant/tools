import json
import boto3
from common.json import Encoder
import sys


class Downloader:
    def __init__(self, region_name: str, endpoint_url=''):
        conf = {
            'region_name': region_name,
        }
        if endpoint_url:
            conf['endpoint_url'] = endpoint_url
        self.client = boto3.client('s3', **conf)
        self.resource = boto3.resource('s3', **conf)

    def _list_objects(self, bucket: str, prefix: str, ct='') -> dict:
        params = {
            'Bucket': bucket,
            'Delimiter': '/',
            'Prefix': prefix,
        }
        if ct:
            params['ContinuationToken']: ct
        return self.client.list_objects_v2(**params)

    def list_objects(self, bucket: str, prefix: str) -> dict:
        cps = set()
        ct = ''
        contents = []
        while True:
            r = self._list_objects(bucket=bucket, prefix=prefix, ct=ct)
            if 'CommonPrefixes' in r:
                cps |= set(x['Prefix'] for x in r['CommonPrefixes'])
            contents.extend(r.get('Contents', []))
            ct = r.get('NextContinuationToken', '')
            if not ct:
                break
        return {
            'Contents': contents,
            'CommonPrefixes': list(cps),
        }

    def _dowload_objects(self, bucket: str, objects: dict = None, dry=False):
        if objects is None:
            objects = {}
        if objects.get('CommonPrefixes'):
            for elem in objects['CommonPrefixes']:
                self._dowload_objects(
                    bucket=bucket,
                    objects=self.list_objects(bucket=bucket, prefix=elem),
                )
            return

        if not objects.get('Contents'):
            print('no contents')
            return

        b = self.resource.Bucket(bucket)
        for elem in objects['Contents']:
            if dry:
                print(json.dumps(elem, cls=Encoder))
                continue
            sys.stdout.write('download {} ...'.format(json.dumps(elem, cls=Encoder)))
            b.download_file(Key=elem['Key'], Filename=elem['Key'].split('/')[-1])
            print('done')

    def downaload_objects(self, bucket: str, prefix: str, dry=False):
        self._dowload_objects(
            bucket=bucket,
            objects=self.list_objects(bucket=bucket, prefix=prefix),
            dry=dry,
        )


if __name__ == '__main__':
    from argparse import ArgumentParser

    p = ArgumentParser(
        description='download objects from s3',
    )

    p.add_argument('--region_name', action='store', type=str, required=True, help='region')
    p.add_argument('--endpoint_url', action='store', default='', help='endpoint')
    p.add_argument('-b', action='store', type=str, required=True, help='bucket')
    p.add_argument('-p', action='store', type=str, default='', help='prefix')
    p.add_argument('--dry', action='store_const', const=True, default=False, help='dry run')

    opt = p.parse_args()

    d = Downloader(
        region_name=opt.region_name,
        endpoint_url=opt.endpoint_url,
    )
    d.downlaod_objects(
        bucket=opt.b,
        prefix=opt.p,
        dry=opt.dry,
    )
