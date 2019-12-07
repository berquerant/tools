import pika
import sys


class Publisher:
    def __init__(self, username: str, password: str, host: str, port: int):
        self.host = host
        self.port = port
        self.credentials = pika.PlainCredentials(
            username=username,
            password=password,
        )

    def basic_publish(self, exchange: str, routing_key: str):
        self.publish(
            exchange=exchange,
            routing_key=routing_key,
            bodies=(x.rstrip() for x in sys.stdin),
        )

    def publish(self, exchange: str, routing_key: str, bodies: iter):
        conn = pika.BlockingConnection(
            pika.ConnectionParameters(
                host=self.host,
                port=self.port,
                credentials=self.credentials,
            ),
        )
        ch = conn.channel()
        try:
            for body in bodies:
                ch.basic_publish(
                    exchange=exchange,
                    routing_key=routing_key,
                    body=body,
                )
        except Exception as e:
            print(e, file=sys.stderr)
        finally:
            ch.close()
            conn.close()


if __name__ == '__main__':
    from argparse import ArgumentParser

    p = ArgumentParser(
        description='rabbitmq publisher',
    )

    p.add_argument('--port', action='store', default=5672, help='port')
    p.add_argument('--host', action='store', type=str, required=True, help='host')
    p.add_argument('-k', action='store', type=str, required=True, help='routing key')
    p.add_argument('-x', action='store', default='', help='exchange')
    p.add_argument('-u', action='store', type=str, required=True, help='user')
    p.add_argument('-p', action='store', type=str, required=True, help='password')

    opt = p.parse_args()

    publisher = Publisher(
        username=opt.u,
        password=opt.p,
        host=opt.host,
        port=opt.port,
    )
    publisher.basic_publish(
        exchange=opt.x,
        routing_key=opt.k,
    )
