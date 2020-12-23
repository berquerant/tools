from graphviz import Digraph
import json
from collections import namedtuple
from copy import copy
import os
from common.json import dumps_for_graphviz


class Destination(namedtuple("Destination", "d")):
    @property
    def id(self) -> str:
        return self.d["id"]

    @property
    def el(self) -> str:
        """edge label"""
        return self.d.get("el", "")


class Row(namedtuple("Row", "d")):
    @property
    def id(self) -> str:
        return self.d["id"]

    @property
    def desc(self) -> dict:
        d = copy(self.d)
        if "to" in d:
            del d["to"]
        return d

    @property
    def to(self) -> list:
        return [Destination(d=x) for x in self.d.get("to", [])]


class JSON2Dot(namedtuple("JSON2Dot", "src dest")):
    def read(self) -> iter:
        for line in self.src:
            yield Row(d=json.loads(line))

    def get(self) -> str:
        g = Digraph(format="pdf", node_attr={
            "shape": "plaintext",
            "style": "solid,filled",
            "width": "1",
            "fontname": "arial",
        })
        self.draw(g, self.read())
        d, f = self.dest.split("/", 1)
        g.render(filename=f, directory=d)
        return os.path.abspath("{}.pdf".format(self.dest))

    def draw(self, g: Digraph, rows: iter):
        node_ids = set()
        edges = []

        for r in rows:
            if r.id not in node_ids:
                node_ids.add(r.id)
                g.node(r.id, label=dumps_for_graphviz(r.desc))
            edges.extend([
                {
                    "from": r.id,
                    "to": d.id,
                    "label": d.el,
                }
                for d in r.to
            ])
        for edge in edges:
            to = edge["to"]
            if to not in node_ids:
                node_ids.add(to)
                g.node(to, label=to)
            g.edge(edge["from"], to, label=edge["label"])


if __name__ == "__main__":
    from argparse import (
        ArgumentParser,
        RawDescriptionHelpFormatter,
    )
    import sys

    p = ArgumentParser(
        description="draw graph from json",
        formatter_class=RawDescriptionHelpFormatter,
        epilog="""
example:
(echo '{"id":"A","to":[{"id":"B","el":"a2b"},{"id":"C","el":"a2c"}]}' ; echo '{"id":"B","to":[{"id":"A"}]}' ; echo '{"id":"C","comment":"Canary","to":[{"id":"A"},{"id":"C","el":"self"}]}' ; echo '{"id":"D","comment":"Doom","star":5}') | python json2dot.py""",
    )

    p.add_argument("-o", action="store", default="tmp/json2dot.out", help="output filename")

    opt = p.parse_args()
    dest = JSON2Dot(src=sys.stdin, dest=opt.o).get()
    print(dest)
