import boto3
import json
from common.json import Encoder


class DynamoStream:
    def __init__(self, region_name: str, endpoint_url=''):
        conf = {
            'region_name: region_name,'
        }
        if endpoint_url:
            conf['endpoint_url'] = endpoint_url
        self.db_client = boto3.client('dynamodb', **conf)
        self.stream_client = boto3.client('dynamodbstreams', **conf)

    def _describe_table(self, table_name: str):
        return self.db_client.describe_table(TableName=table_name)

    def _describe_stream(self, stream_arn: str):
        return self.stream_client.describe_stream(StreamArn=stream_arn)

    def _get_shard_iterator(self, stream_arn: str, shard_id: str):
        return self.stream_client.get_shard_iterator(
            StreamArn=stream_arn,
            ShardId=shard_id,
            ShardIteratorType='LATEST',
        )

    def _get_records(self, shard_iterator: str):
        return self.stream_client.get_records(ShardIterator=shard_iterator)

    def get_records(self, table_name: str)-> iter:
        stream_arn = self._describe_table(table_name).get('Table', {}).get('LatestStreamArn', '')
        if not stream_arn:
            return []

        shard_ids = (
            x['ShardId'] for x in
            self._describe_stream(stream_arn).get('StreamDescription', {}).get('Shards', [])
            if 'ShardId' in x
        )
        shard_iterators = (
            self._get_shard_iterator(stream_arn=stream_arn, shard_id=x).get('ShardIterator', '')
            for x in shard_ids
        )

        for si in shard_iterators:
            csi = si
            while csi:
                res = self._get_records(csi)
                if res.get('Records'):
                    yield res
                csi = res.get('NextShardIterator', '')


if __name__ == '__main__':
    from argparse import ArgumentParser

    p = ArgumentParser(
        description='yield dynamodb stream',
    )
    p.add_argument('--region_name', action='store', type=str, required=True)
    p.add_argument('--endpoint_url', action='store', type=str)
    p.add_argument('--table_name', action='store', type=str, required=True)

    opt = p.parse_args()
    ds = DynamoStream(
        region_name=opt.region_name,
        endpoint_url=opt.endpoint_url,
    )
    for x in ds.get_records(opt.table_name):
        print(json.dumps(x, cls=Encoder))
