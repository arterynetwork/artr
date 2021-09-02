from datetime import timedelta

from .config import Config


def height_to_time(height: int, config: Config) -> str:
    try:
        return (config.get_genesis_time() + timedelta(seconds=(height-config.initial_height)*30 // config.time_quotient)).isoformat() + "Z"
    except OverflowError:
        return "9999-12-31T23:59:59.999999999Z"


def modulo_to_time(modulo: int, config: Config) -> str:
    return height_to_time(config.initial_height + (modulo - config.initial_height) % 2880, config)


_STATUSES_MAPPING = {
    0: "STATUS_UNSPECIFIED",
    1: "STATUS_LUCKY",
    2: "STATUS_LEADER",
    3: "STATUS_MASTER",
    4: "STATUS_CHAMPION",
    5: "STATUS_BUSINESSMAN",
    6: "STATUS_PROFESSIONAL",
    7: "STATUS_TOP_LEADER",
    8: "STATUS_HERO",
    9: "STATUS_ABSOLUTE_CHAMPION",
}


def patch_status(status: int) -> str:
    return _STATUSES_MAPPING[status]
