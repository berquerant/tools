from kafka import KafkaConsumer
import sys


class Consumer:
    def __init__(self, servers: list, group_id: str = None):
        self.servers = servers
        self.group_id = group_id

    def connect(self, topics: list = None, pattern: str = None) -> KafkaConsumer:
        if topics is None and pattern is None:
            raise Exception("topics or pattern are required")

        def deserializer(x: bytes) -> str:
            return x.decode("utf-8")

        c = KafkaConsumer(
            bootstrap_servers=self.servers,
            key_deserializer=deserializer,
            value_deserializer=deserializer,
            group_id=self.group_id,
        )
        if not c.bootstrap_connected():
            raise Exception("cannot connect to bootstrap")
        if topics:
            c.subscribe(topics=topics)
        else:
            c.subscribe(pattern=pattern)
        return c

    @staticmethod
    def basic_consumer(msg):
        print("{} {} {} {} {}".format(
            msg.key, msg.timestamp, msg.partition, msg.offset, msg.value))

    def consume(self, topics: list = None, pattern: str = None):
        conn = self.connect(topics=topics, pattern=pattern)
        try:
            for msg in conn:
                self.basic_consumer(msg)
        except Exception as e:
            print(e, file=sys.stderr)
        finally:
            conn.close()


if __name__ == "__main__":
    from argparse import ArgumentParser

    p = ArgumentParser(description="kafka consumer")
    p.add_argument("servers", metavar="EP", type=str, nargs="*", help="bootstrap servers")
    p.add_argument("-g", action="store", type=str, default=None, help="group id")
    p.add_argument("-t", action="store", type=str, required=True, help="choose topics, csv or regex")
    opt = p.parse_args()

    cc = {}
    if "," in opt.t:
        cc["topics"] = opt.t.split(",")
    else:
        cc["pattern"] = opt.t
    Consumer(servers=opt.servers, group_id=opt.g).consume(**cc)
