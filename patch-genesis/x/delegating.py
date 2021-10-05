import operator
from typing import Dict, Optional

from ..config import Config
from ..util import height_to_time, modulo_to_time


def patch(state: Optional[Dict], config: Config) -> Dict:
    state = state or {}
    accounts: Dict = {}
    params: Dict = {}

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

    for key, val in (state.get("params", None) or {}).items():
        if key == "revoke_period":
            val = int(val) // 2880
        params[key] = val

    return {
        "params":   params,
        "accounts": sorted(accounts.values(), key=operator.itemgetter("address"))
    }
