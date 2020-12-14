from graphviz import Digraph
from csv import reader
from collections import namedtuple
import os


class Row(namedtuple("Row", "parent children")):
    pass


class CSV2Dot(namedtuple("CSV2Dot", "src dest")):
    def read(self) -> iter:
        r = reader(self.src)
        for row in r:
            yield Row(parent=row[0], children=row[1:])

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

        def new_node(nid: str):
            node_ids.add(nid)
            g.node(nid, label=nid)

        for r in rows:
            if r.parent not in node_ids:
                new_node(r.parent)
            for c in r.children:
                if c not in node_ids:
                    new_node(c)
                g.edge(r.parent, c)


if __name__ == "__main__":
    from argparse import (
        ArgumentParser,
        RawDescriptionHelpFormatter,
    )
    import sys

    p = ArgumentParser(
        description="draw graph from csv",
        formatter_class=RawDescriptionHelpFormatter,
        epilog="""
example:
(echo A,B,C ; echo B,A ; echo C,B ; echo D) | python csv2dot.py""",
    )

    p.add_argument("-o", action="store", default="tmp/csv2dot.out", help="output filename")

    opt = p.parse_args()
    dest = CSV2Dot(src=sys.stdin, dest=opt.o).get()
    print(dest)
