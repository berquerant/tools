from html.parser import HTMLParser
from html.entities import name2codepoint
from enum import Enum
import json


class Tag(Enum):
    START_TAG = 1
    END_TAG = 2
    STARTEND_TAG = 3
    DATA = 4
    ENTITY_REF = 5
    CHAR_REF = 6
    COMMENT = 7
    DECL = 8
    PI = 9
    UNKNOWN = 100


class Dumper(HTMLParser):
    def dump(self, xs: list):
        print("\t".join(str(x) for x in xs), flush=True)

    def log(self, xs: list):
        r, c = self.getpos()
        self.dump([r, c, *xs])

    def translate_data(self, x: str):
        return "\\n".join(x.splitlines())

    def translate_attrs(self, attrs: list):
        return json.dumps(attrs, separators=(",", ":"))

    def handle_starttag(self, tag: str, attrs: list):
        self.log([Tag.START_TAG, tag, self.translate_attrs(attrs)])

    def handle_endtag(self, tag: str):
        self.log([Tag.END_TAG, tag])

    def handle_startendtag(self, tag: str, attrs: list):
        self.log([Tag.STARTEND_TAG, tag, self.translate_attrs(attrs)])

    def handle_data(self, data: str):
        self.log([Tag.DATA, self.translate_data(data)])

    def handle_entityref(self, name: str):
        self.log([Tag.ENTITY_REF, chr(name2codepoint[name])])

    def handle_charref(self, name: str):
        i = int(name[1:], 16) if name.startswith("x") else int(name)
        self.log([Tag.CHAR_REF, chr(i)])

    def handle_comment(self, data: str):
        self.log([Tag.COMMENT, self.translate_data(data)])

    def handle_decl(self, decl: str):
        self.log([Tag.DECL, decl])

    def handle_pi(self, data: str):
        self.log([Tag.PI, data])

    def unknown_decl(self, data: str):
        self.log([Tag.UNKNOWN, self.translate_data(data)])


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
