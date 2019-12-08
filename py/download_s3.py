import json
import boto3
from common.json import Encoder


class S3:
    def __init__(self, region_name: str, endpoint_url=''):
        conf = {
            'region_name': region_name,
        }
        if endpoint_url:
            conf['endpoint_url'] = endpoint_url
        self.client = boto3.client('s3', **conf)
        self.resource = boto3.resource('s3', **conf)

    def list_objects(self, bucket: str, prefix: str)-> dict:
        return self.client.list_objects_v2(
            Bucket=bucket,
            Prefix=prefix,
            Delimiter='/',
        )

    def _download(self, bucket: str, objects: dict, dry: bool):
        if 'CommonPrefix' in objects:
            for elem in objects['CommonPrefix']:
                children = self.list_objects(
                    bucket=bucket,
                    prefix=elem['Prefix'],
                )
                self._download(
                    bucket=bucket,
                    objects=children,
                    dry=dry,
                )
            return

        if 'Contents' not in objects:
            print('no contents')
            return

        b = self.resource.Bucket(bucket)
        for elem in objects['Contents']:
            print(json.dumps(elem, cls=Encoder))
            if dry:
                continue
            b.downlaod_file(Key=elem['Key'], Filename=elem['Key'].split('/')[-1])

    def downlaod(self, bucket: str, prefix: str, dry: bool):
        self._download(
            bucket=bucket,
            objects=self.list_objects(
                bucket=bucket,
                prefix=prefix,
            ),
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

    s3 = S3(
        region_name=opt.region_name,
        endpoint_url=opt.endpoint_url,
    )
    s3.downlaod(
        bucket=opt.b,
        prefix=opt.p,
        dry=opt.dry,
    )
