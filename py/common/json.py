from decimal import Decimal
from datetime import (
    date,
    datetime,
)
from json import (
    JSONEncoder,
    dumps,
)


class Encoder(JSONEncoder):
    """json encoding utility"""

    def default(self, obj):
        if isinstance(obj, (datetime, date)):
            return obj.isoformat()
        if isinstance(obj, Decimal):
            f = float(obj)
            if f.is_integer():
                return int(f)
            return f
        if isinstance(obj, set):
            return list(obj)
        if isinstance(obj, bytes):
            return str(obj)
        return JSONEncoder.default(self, obj)


def dumps_compact(obj) -> str:
    return dumps(obj, cls=Encoder, separators=(',', ':'))


def dumps_for_graphviz(obj) -> str:
    return dumps(obj, sort_keys=True, indent=" ").\
        replace("\n", "\\l").replace("{", "\\{").replace("}", "\\}") + "\\l"
