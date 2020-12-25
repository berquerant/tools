from common.json import dumps_for_graphviz
from abc import (
    ABC,
    abstractmethod,
)
from graphviz import Digraph
from os import path
from collections import namedtuple
import csv
from uuid import uuid4
import json
from copy import copy


class Dot(ABC):
    def new_digraph(self, fmt: str) -> Digraph:
        return Digraph(format=fmt, node_attr={
            "shape": "plaintext",
            "style": "solid,filled",
            "width": "1",
            "fontname": "arial",
        })

    @staticmethod
    def find_format(dest: str) -> str:
        x = path.basename(path.abspath(dest)).split(".")
        if len(x) > 1 and x[-1]:
            return x[-1]
        raise Exception("output format not found in {}".format(dest))

    def run(self, src, dest: str) -> str:
        g = self.new_digraph(self.find_format(dest))
        self.draw(g, src)
        p = path.abspath(dest)
        fs = path.basename(p).split(".")
        f = ".".join(fs[:len(fs) - 1])
        g.render(directory=path.dirname(p), filename=f)
        return p

    @abstractmethod
    def draw(self, g: Digraph, src):
        """src: file"""


class CSV2Dot(Dot):
    def draw(self, g: Digraph, src):
        Row = namedtuple("Row", "parent children")

        def read() -> iter:
            for x in csv.reader(src):
                yield Row(parent=x[0], children=x[1:])

        nids = set()

        def new_node(nid: str):
            if nid not in nids:
                return
            nids.add(nid)
            g.node(nid, label=nid)

        for r in read():
            new_node(r.parent)
            for c in r.children:
                new_node(c)
                g.edge(r.parent, c)


class JSONTreeDot(Dot):
    def __init__(self, children: list):
        self.children = children

    def new_nid() -> str:
        return str(uuid4())

    def draw(self, g: Digraph, src):
        root = json.loads(src.read())
        if not isinstance(root, dict):
            raise Exception("root must be object")

        def new_nid() -> str:
            return str(uuid4())

        def _draw(x, nid: str, edge_name="", parent_id=""):
            if not isinstance(x, dict):
                g.node(nid, label=dumps_for_graphviz(x))
                if parent_id:
                    g.edge(parent_id, nid, label=edge_name)
                return

            cs = {k: v for k, v in x.items() if k in self.children}
            for k in cs.keys():
                del x[k]
            g.node(nid, label=dumps_for_graphviz(x))
            if parent_id:
                g.edge(parent_id, nid, label=edge_name)
            for k, v in cs.items():
                _draw(x=v, nid=new_nid(), edge_name=k, parent_id=nid)

        _draw(x=root, nid=new_nid())


class JSON2Dot(Dot):
    def draw(self, g: Digraph, src):
        class Dest(namedtuple("Dest", "d")):
            @property
            def id(self) -> str:
                return self.d["id"]

            @property
            def el(self) -> str:
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
                return [Dest(d=x) for x in self.d.get("to", [])]

        def read() -> iter:
            for l in src:
                yield Row(d=json.loads(l))

        nids = set()
        edges = []
        for r in read():
            if r.id not in nids:
                nids.add(r.id)
                g.node(r.id, label=dumps_for_graphviz(r.desc))
            edges.extend([
                {
                    "from": r.id,
                    "to": d.id,
                    "label": d.el,
                } for d in r.to
            ])
        for edge in edges:
            to = edge["to"]
            if to not in nids:
                nids.add(to)
                g.node(to, label=to)
            g.edge(edge["from"], to, label=edge["label"])


if __name__ == "__main__":
    import sys
    from argparse import (
        ArgumentParser,
        RawDescriptionHelpFormatter,
    )
    p = ArgumentParser("dot", formatter_class=RawDescriptionHelpFormatter, epilog="""examples:
(echo A,B,C ; echo B,A ; echo C,B ; echo D) | python dot.py -o out.pdf csv2dot

echo '{"n":"N1","lh":{"n":"N2","lh":{"n":"N3"},"rh":{"n":"N4"}},"rh":{"n":"N5","lh":{"n":"N6"},"rh":{"n":"N7","t":"true","rh":["M1","M2"]}}}' | python dot.py -o out.pdf jsontreedot lh rh

(echo '{"id":"A","to":[{"id":"B","el":"a2b"},{"id":"C","el":"a2c"}]}' ; echo '{"id":"B","to":[{"id":"A"}]}' ; echo '{"id":"C","comment":"Canary","to":[{"id":"A"},{"id":"C","el":"self"}]}' ; echo '{"id":"D","comment":"Doom","star":5}') | python dot.py -o out.pdf json2dot
""")
    p.add_argument("-o", action="store", type=str, required=True, help="output filename")
    sp = p.add_subparsers(dest="cmd")

    c2d = sp.add_parser("csv2dot", help="""graph from csv""")

    jtd = sp.add_parser("jsontreedot", help="""graph from json as tree.""")
    jtd.add_argument("children", metavar="C", type=str, nargs="+", help="child keys")

    j2d = sp.add_parser("json2dot", help="""graph from json.""")

    args = p.parse_args()

    def get_runner() -> Dot:
        cmd = args.cmd
        if cmd == "csv2dot":
            return CSV2Dot()
        if cmd == "jsontreedot":
            return JSONTreeDot(children=args.children)
        if cmd == "json2dot":
            return JSON2Dot()
        raise Exception("unknown command: {}".format(cmd))

    runner = get_runner()
    print(runner.run(sys.stdin, args.o))
