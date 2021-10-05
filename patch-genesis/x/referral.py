from typing import Dict, Optional

from ..config import Config
from ..util import height_to_time, patch_status


def patch(state: Optional[Dict], config: Config) -> Dict:
    state = state or {}

    result = {
        "params":             {},
        "top_level_accounts": state.get("top_level_accounts", []),
        "other_accounts":     state.get("other_accounts", [])
    }
    for k, v in state.get("params", {}).items():
        if k == "transition_cost":
            k = "transition_price"
        result["params"][k] = v

    result["transitions"] = [
        {
            "subject":     x["subj"],
            "destination": x["dest"]
        } for x in state.get("transitions", None) or []
    ]

    result["compressions"] = [
        {
            "account": x["account"],
            "time":    height_to_time(int(x["height"]), config)
        } for x in state.get("compression", None) or []
    ]

    result["downgrades"] = [
        {
            "account": x["account"],
            "current": patch_status(x["current"]),
            "time":    height_to_time(int(x.pop("height")), config)
        } for x in state.get("downgrade", None) or []
    ]

    result["banished_accounts"] = [
        {
            "account":         x["acc"],
            "former_referrer": x["ref"],
        } for x in state.get("banished_accounts", None) or []
    ]

    result["never_paid"] = state.get("never_paid", None) or []

    return result
