from html.parser import HTMLParser
from html.entities import name2codepoint
import json


class Dumper(HTMLParser):
    def log(self, xs: list):
        r, c = self.getpos()
        print("{}\t{}\t{}".format(r, c, "\t".join(xs)), flush=True)

    def __visualize_nl(self, x: str) -> str:
        return "\\n".join(x.split())

    def __attrs2str(self, attrs: list) -> str:
        return json.dumps(attrs, separators=(",", ":"))

    def handle_starttag(self, tag: str, attrs: list):
        self.log(["startag", tag, self.__attrs2str(attrs)])

    def handle_endtag(self, tag: str):
        self.log(["endtag", tag])

    def handle_startendtag(self, tag: str, attrs: list):
        self.log(["startendtag", tag, self.__attrs2str(attrs)])

    def handle_data(self, data: str):
        self.log(["data", self.__visualize_nl(data)])

    def handle_entityref(self, name: str):
        self.log(["entityref", chr(name2codepoint[name])])

    def handle_charref(self, name: str):
        i = int(name[1:], 16) if name.startswith("x") else int(name)
        self.log(["charref", chr(i)])

    def handle_comment(self, data: str):
        self.log(["comment", self.__visualize_nl(data)])

    def handle_decl(self, decl: str):
        self.log(["decl", decl])

    def handle_pi(self, data: str):
        self.log(["pi", data])

    def unknown_decl(self, data: str):
        self.log(["unknown", self.__visualize_nl(data)])


if __name__ == "__main__":
    import sys
    d = Dumper()
    try:
        for l in sys.stdin:
            d.feed(l)
        d.close()
    except (BrokenPipeError, IOError):
        pass
    sys.stderr.close()
