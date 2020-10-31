def split_list(xs: list, n: int) -> list:
    return [xs[i:i + n] for i in range(0, len(xs), n)]
