from kafka import KafkaProducer
from uuid import uuid4
import sys


class Producer:
    def __init__(self, servers: list):
        self.servers = servers

    @staticmethod
    def request_id() -> str:
        return str(uuid4())

    def connect(self) -> KafkaProducer:
        def serializer(x: str) -> bytes:
            return x.encode("utf-8")

        p = KafkaProducer(
            bootstrap_servers=self.servers,
            key_serializer=serializer,
            value_serializer=serializer,
        )
        if not p.bootstrap_connected():
            raise Exception("cannot connect to bootstrap")
        return p

    def basic_produce(self, topic: str):
        self.produce(topic=topic, bodies=(x.rstrip() for x in sys.stdin))

    def produce(self, topic: str, bodies: iter):
        conn = self.connect()
        try:
            for body in bodies:
                rid = self.request_id()
                rmd = conn.send(topic, key=rid, value=body).get()
                print("{} {} {} {}".format(
                    rid, rmd.timestamp, rmd.partition, rmd.offset, body))
            conn.flush()
        except Exception as e:
            print(e, file=sys.stderr)


if __name__ == "__main__":
    from argparse import ArgumentParser

    p = ArgumentParser(description="kafka producer")
    p.add_argument("servers", metavar="EP", type=str, nargs="*", help="bootstrap servers")
    p.add_argument("-t", action="store", type=str, required=True, help="topic")

    opt = p.parse_args()
    Producer(opt.servers).basic_produce(opt.t)
