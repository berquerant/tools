import copy
from html.parser import HTMLParser
from collections import namedtuple


NO_END_TAGS = [
    "br", "img", "hr", "meta", "input",
    "embed", "area", "base", "col", "keygen",
    "link", "param", "source",
]
# OMITTABLE_END_TAGS = [
#     "p", "dt", "dd", "li", "option",
#     "thead", "tfoot", "th", "tr", "td",
#     "rt", "rp", "optgroup", "caption",
# ]

TracerCell = namedtuple("TracerCell", "tag attrs")
TracerCellAttr = namedtuple("TracerCellAttr", "name value")


class Data(namedtuple("Data", "cells content")):
    def asdict(self) -> dict:
        return {
            "content": self.content,
            "cells": [
                {
                    "tag": cell.tag,
                    "attrs": [
                        attr._asdict() for attr in cell.attrs
                    ],
                } for cell in self.cells
            ],
        }


class Tracer:
    def __init__(self):
        self.__d = []

    def __len__(self):
        return len(self.__d)

    def put(self, x: TracerCell):
        self.__d.append(x)

    def pop(self) -> TracerCell:
        return self.__d.pop()

    @property
    def top(self) -> TracerCell:
        return self.__d[-1] if self.__d else None

    @property
    def data(self) -> list:
        return copy.copy(self.__d)


class Extractor(HTMLParser):
    def __init__(self):
        super().__init__()
        self.__tracer = Tracer()
        self.__data = []

    @property
    def data(self) -> list:
        return self.__data

    def __save(self, cells: list, content=""):
        self.__data.append(Data(cells=cells, content=content))

    def __pop_no_end_tag(self):
        if self.__tracer.top and self.__tracer.top.tag in NO_END_TAGS:
            self.__tracer.pop()

    def handle_starttag(self, tag, attrs):
        self.__pop_no_end_tag()
        tc = TracerCell(
            tag=tag,
            attrs=[TracerCellAttr(name=n, value=v) for n, v in attrs],
        )
        self.__tracer.put(tc)
        self.__save(cells=self.__tracer.data)

    def handle_data(self, data):
        self.__save(cells=self.__tracer.data, content=data)

    def handle_endtag(self, tag):
        self.__pop_no_end_tag()
        self.__tracer.pop()

    def extract(self, data: iter):
        for x in data:
            self.feed(x)
        self.close()


if __name__ == "__main__":
    import sys
    import json
    p = Extractor()
    p.extract(x.rstrip() for x in sys.stdin)
    for x in p.data:
        print(json.dumps(x.asdict()))
