from graphviz import Digraph
import json
from collections import namedtuple
from uuid import uuid4
from common.json import dumps_for_graphviz
import os


class JSONTree(namedtuple("JSONTree", "src children dest")):
    def get(self) -> str:
        x = json.loads(self.src)
        if not isinstance(x, dict):
            raise Exception("not object")
        g = Digraph(format="pdf", node_attr={
            "shape": "plaintext",
            "style": "solid,filled",
            "width": "1",
            "fontname": "arial",
        })
        self.draw(g=g, x=x, nid=self.new_nid())
        d, f = self.dest.split("/", 1)
        g.render(filename=f, directory=d)
        return os.path.abspath("{}.pdf".format(self.dest))

    def new_nid(self) -> str:
        return str(uuid4())

    def draw(self, g: Digraph, x, nid, edge_name="", parent_id=""):
        if not isinstance(x, dict):
            g.node(str(nid), label=dumps_for_graphviz(x))
            if parent_id:
                g.edge(str(parent_id), str(nid), label=edge_name)
            return

        cs = {k: v for k, v in x.items() if k in self.children}
        for k in cs.keys():
            del x[k]
        g.node(str(nid), label=dumps_for_graphviz(x))
        if parent_id:
            g.edge(str(parent_id), str(nid), label=edge_name)
        for k in sorted(cs.keys()):
            self.draw(g=g, edge_name=k, x=cs[k], parent_id=nid, nid=self.new_nid())


if __name__ == "__main__":
    from argparse import (
        ArgumentParser,
        RawDescriptionHelpFormatter,
    )
    import sys

    p = ArgumentParser(
        description="draw tree from json",
        formatter_class=RawDescriptionHelpFormatter,
        epilog="""
example:
echo '{"n":"N1","lh":{"n":"N2","lh":{"n":"N3"},"rh":{"n":"N4"}},"rh":{"n":"N5","lh":{"n":"N6"},"rh":{"n":"N7","t":"true","rh":["M1","M2"]}}}' | python jsontree.py lh rh""",
    )

    p.add_argument("-o", action="store", default="tmp/jsontree.out", help="output filename")
    p.add_argument("children", metavar="C", type=str, nargs="+", help="children keys")

    opt = p.parse_args()
    src = sys.stdin.read()
    dest = JSONTree(src=src, children=opt.children, dest=opt.o).get()
    print(dest)
