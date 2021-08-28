import operator
from typing import Dict, Optional

from ..config import Config
from ..util import height_to_time, modulo_to_time


def patch(state: Optional[Dict], config: Config) -> Dict:
    state = state or {}
    accounts: Dict = {}

    for x in state.get("revoking", None) or []:
        account = accounts.get(x["account"])
        if account is None:
            accounts[x["account"]] = account = {
                "address":  x["account"],
                "requests": []
            }
        account["requests"].append({
            "time":   height_to_time(int(x["height"]), config),
            "amount": x["amount"]
        })

    for val in state.get("clusters", None) or []:
        t: str = modulo_to_time(val["modulo"], config)
        for x in val["accounts"]:
            acc = accounts.get(x)
            if acc is None:
                accounts[x] = acc = {
                    "address": x,
                }
            acc["next_accrue"] = t

    return {
        "params":   state.get("params"),
        "accounts": sorted(accounts.values(), key=operator.itemgetter("address"))
    }
