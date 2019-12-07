import pika
import sys


class Consumer:
    def __init__(self, username: str, password: str, host: str, port: int):
        self.host = host
        self.port = port
        self.credentials = pika.PlainCredentials(
            username=username,
            password=password,
        )

    @staticmethod
    def basic_consumer():
        def inner(channel, method, header, body):
            print(body.decode('utf-8'), flush=True)
            channel.basic_ack(delivery_tag=method.delivery_tag)
        return inner

    def consume(self, exchange: str, queue: str, callback: callable):
        conn = pika.BlockingConnection(
            pika.ConnectionParameters(
                host=self.host,
                port=self.port,
                credentials=self.credentials,
            ),
        )
        ch = conn.channel()
        ch.queue_declare(queue=queue, durable=True)
        if exchange:
            ch.queue_bind(
                queue=queue,
                exchange=exchange,
                routing_key='#',
            )

        try:
            ch.basic_qos(prefetch_count=0)
            ch.basic_consume(queue=queue, on_message_callback=callback)
            ch.start_consuming()
        except Exception as e:
            print(e, file=sys.stderr)
        finally:
            ch.stop_consuming()
            ch.close()
            conn.close()


if __name__ == '__main__':
    from argparse import ArgumentParser

    p = ArgumentParser(
        description='rabbitmq consumer',
    )

    p.add_argument('--port', action='store', default=5672, help='port')
    p.add_argument('--host', action='store', type=str, required=True, help='host')
    p.add_argument('-q', action='store', type=str, required=True, help='queue')
    p.add_argument('-x', action='store', default='', help='exchange')
    p.add_argument('-u', action='store', type=str, required=True, help='user')
    p.add_argument('-p', action='store', type=str, required=True, help='password')

    opt = p.parse_args()

    consumer = Consumer(
        username=opt.u,
        password=opt.p,
        host=opt.host,
        port=opt.port,
    )
    consumer.consume(
        exchange=opt.x,
        queue=opt.q,
        callback=Consumer.basic_consumer(),
    )
