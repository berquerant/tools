from graphviz import Digraph
import json
from collections import namedtuple
from uuid import uuid4


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

    def new_nid(self) -> str:
        return str(uuid4())

    def json_dumps(self, x) -> str:
        return json.dumps(x, sort_keys=True, indent=" ").\
            replace("\n", "\\l").replace("{", "\\{").replace("}", "\\}") + "\\l"

    def draw(self, g: Digraph, x, nid, edge_name="", parent_id=""):
        if not isinstance(x, dict):
            g.node(str(nid), label=self.json_dumps(x))
            if parent_id:
                g.edge(str(parent_id), str(nid), label=edge_name)
            return

        cs = {k: v for k, v in x.items() if k in self.children}
        for k in cs.keys():
            del x[k]
        g.node(str(nid), label=self.json_dumps(x))
        if parent_id:
            g.edge(str(parent_id), str(nid), label=edge_name)
        for k in sorted(cs.keys()):
            self.draw(g=g, edge_name=k, x=cs[k], parent_id=nid, nid=self.new_nid())


if __name__ == "__main__":
    from argparse import ArgumentParser
    import sys

    p = ArgumentParser(
        description="draw tree from json",
    )

    p.add_argument("-o", action="store", default="tmp/jsontree.out", help="output filename")
    p.add_argument("children", metavar="C", type=str, nargs="+", help="children keys")

    opt = p.parse_args()
    src = sys.stdin.read()
    JSONTree(src=src, children=opt.children, dest=opt.o).get()
