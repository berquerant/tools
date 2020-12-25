from html.parser import HTMLParser
from abc import (
    ABC,
    abstractmethod,
)
from enum import Enum
import json
from collections import OrderedDict


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


class HTMLDumper(HTMLParser, ABC):
    def __init__(self):
        super().__init__(convert_charrefs=True)

    @abstractmethod
    def translate_data(self, x: str):
        pass

    @abstractmethod
    def translate_attrs(self, attrs: list):
        pass

    @abstractmethod
    def log(self, x: dict):
        pass

    def gen_log(self, kind: Tag) -> OrderedDict:
        return OrderedDict(kind=str(kind))

    def handle_starttag(self, tag: str, attrs: list):
        d = self.gen_log(Tag.START_TAG)
        d["tag"] = tag
        d["attrs"] = self.translate_attrs(attrs)
        self.log(d)

    def handle_endtag(self, tag: str):
        d = self.gen_log(Tag.END_TAG)
        d["tag"] = tag
        self.log(d)

    def handle_startendtag(self, tag: str, attrs: list):
        d = self.gen_log(Tag.STARTEND_TAG)
        d["tag"] = tag
        d["attrs"] = self.translate_attrs(attrs)
        self.log(d)

    def handle_data(self, data: str):
        d = self.gen_log(Tag.DATA)
        d["data"] = self.translate_data(data)
        self.log(d)

    def handle_comment(self, data: str):
        d = self.gen_log(Tag.COMMENT)
        d["data"] = self.translate_data(data)
        self.log(d)

    def handle_decl(self, decl: str):
        d = self.gen_log(Tag.DECL)
        d["data"] = self.translate_data(decl)
        self.log(d)

    def handle_pi(self, data: str):
        d = self.gen_log(Tag.PI)
        d["data"] = self.translate_data(data)
        self.log(d)

    def unknown_decl(self, data: str):
        d = self.gen_log(Tag.UNKNOWN)
        d["data"] = self.translate_data(data)
        self.log(d)


class HTMLJSONDumper(HTMLDumper):
    def translate_data(self, x: str):
        return "\\n".join(x.splitlines())

    def translate_attrs(self, attrs: list):
        return attrs

    def log(self, x: dict):
        print(json.dumps(x, separators=(",", ":")))


class HTMLTSVDumper(HTMLDumper):
    def translate_data(self, x: str):
        return "\\n".join(x.splitlines())

    def translate_attrs(self, attrs: list):
        return json.dumps(attrs, separators=(",", ":"))

    def log(self, x: dict):
        print("\t".join(str(v) for v in x.values()))


if __name__ == "__main__":
    import sys
    from argparse import ArgumentParser

    p = ArgumentParser("html_dumper")
    p.add_argument("-f", action="store", default="json", choices=["json", "tsv"], help="format")
    args = p.parse_args()

    def dumper(f: str) -> HTMLDumper:
        if f == "json":
            return HTMLJSONDumper()
        if f == "tsv":
            return HTMLTSVDumper()
        raise Exception("unknown dumper: {}".format(f))

    d = dumper(args.f)
    try:
        for l in sys.stdin:
            d.feed(l)
    except BrokenPipeError:
        pass
    finally:
        d.close()
