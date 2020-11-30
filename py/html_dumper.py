from html.parser import HTMLParser
from html.entities import name2codepoint
import json


class Dumper(HTMLParser):
    def dump(self, xs: list):
        print("\t".join(str(x) for x in xs), flush=True)

    def log(self, xs: list):
        r, c = self.getpos()
        self.dump([r, c, *xs])

    def translate_data(self, x: str):
        return "\\n".join(x.split())

    def translate_attrs(self, attrs: list):
        return json.dumps(attrs, separators=(",", ":"))

    def handle_starttag(self, tag: str, attrs: list):
        self.log(["startag", tag, self.translate_attrs(attrs)])

    def handle_endtag(self, tag: str):
        self.log(["endtag", tag])

    def handle_startendtag(self, tag: str, attrs: list):
        self.log(["startendtag", tag, self.translate_attrs(attrs)])

    def handle_data(self, data: str):
        self.log(["data", self.translate_data(data)])

    def handle_entityref(self, name: str):
        self.log(["entityref", chr(name2codepoint[name])])

    def handle_charref(self, name: str):
        i = int(name[1:], 16) if name.startswith("x") else int(name)
        self.log(["charref", chr(i)])

    def handle_comment(self, data: str):
        self.log(["comment", self.translate_data(data)])

    def handle_decl(self, decl: str):
        self.log(["decl", decl])

    def handle_pi(self, data: str):
        self.log(["pi", data])

    def unknown_decl(self, data: str):
        self.log(["unknown", self.translate_data(data)])


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
