import base64
import bech32
from typing import Dict, Optional

from ..config import Config
from ..util import height_to_time, modulo_to_time


def _patch_acc_address(data: str) -> str:
    bz: bytes = base64.standard_b64decode(data)
    acc_addr: str = bech32.bech32_encode("artr", bech32.convertbits(bz, 8, 5))
    return base64.standard_b64encode(acc_addr.encode()).decode()


def _patch_schedule_task(task: Dict, height: int, config: Config) -> Dict:
    return {
        "handler_name": "profile/refresh" if task["handler_name"] == "subscription/refresh" else task["handler_name"],
        "data":         _patch_acc_address(task["data"]) if task["handler_name"] in {
                            "referral/compression", "referral/downgrade", "referral/transition-timeout"
                        } else task["data"],
        "time":         height_to_time(height, config)
    }


def patch(state: Optional[Dict], delegating: Optional[Dict], config: Config) -> Dict:
    tasks = [
        _patch_schedule_task(task, int(x["height"]), config)
        for x in state.get("tasks") or []
        for task in x.get("Schedule", [])
    ]
    if delegating:
        for cluster in delegating.get("clusters", None) or []:
            t: str = modulo_to_time(cluster["modulo"], config)
            tasks.extend({
                "handler_name": "delegating/accrue",
                "time":         t,
                "data":         base64.standard_b64encode(bytes(bech32.convertbits(bech32.bech32_decode(acc)[1], 5, 8))).decode()
            } for acc in cluster["accounts"])

    return {
        "params": {
            "day_nanos": str(86400000000000 // config.time_quotient)
        },
        "tasks": tasks
    }
